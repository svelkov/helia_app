// global/vars.go
package global

import (
	"fmt"
	"sync"
	"time"
)

var (
	headerLocks     sync.Map // Global map for managing resource locks by ID
	entityLockTable sync.Map // Map[int64]*EntityLock - tracks who has what locked
)

// EntityLock represents an active lock on an entity (nalog, partner, etc.)
type EntityLock struct {
	EntityID   int64
	EntityType string // e.g., "nalog", "partner", "user"
	Username   string
	SessionID  string
	LockedAt   time.Time
	LastAccess time.Time
	mu         sync.RWMutex
}

// GetHeaderLock retrieves or creates a mutex for the given resource ID
func GetHeaderLock(resourceID interface{}) *sync.Mutex {
	lock, _ := headerLocks.LoadOrStore(resourceID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// TryLockEntity attempts to acquire a lock on an entity
// Returns: (success, existingLock, error)
func TryLockEntity(entityID int64, entityType, username, sessionID string) (bool, *EntityLock, error) {
	now := time.Now()

	// Create new lock object
	newLock := &EntityLock{
		EntityID:   entityID,
		EntityType: entityType,
		Username:   username,
		SessionID:  sessionID,
		LockedAt:   now,
		LastAccess: now,
	}

	// Try to atomically store the lock
	actual, loaded := entityLockTable.LoadOrStore(entityID, newLock)

	if !loaded {
		// Successfully acquired new lock
		return true, newLock, nil
	}

	// Lock exists, check if it's ours or stale
	existingLock := actual.(*EntityLock)
	existingLock.mu.Lock()
	defer existingLock.mu.Unlock()

	// Same session? Refresh the lock
	if existingLock.SessionID == sessionID {
		existingLock.LastAccess = now
		return true, existingLock, nil
	}

	// Check if lock is stale (no activity for 15 minutes)
	if time.Since(existingLock.LastAccess) > 15*time.Minute {
		// Lock is stale, replace it
		entityLockTable.Store(entityID, newLock)
		return true, newLock, nil
	}

	// Lock is held by another user
	return false, existingLock, nil
}

// RefreshLock updates the last access time for a lock
func RefreshLock(entityID int64, sessionID string) bool {
	if val, ok := entityLockTable.Load(entityID); ok {
		lock := val.(*EntityLock)
		lock.mu.Lock()
		defer lock.mu.Unlock()

		if lock.SessionID == sessionID {
			lock.LastAccess = time.Now()
			return true
		}
	}
	return false
}

// UnlockEntity releases a lock if owned by the session
func UnlockEntity(entityID int64, sessionID string) bool {
	if val, ok := entityLockTable.Load(entityID); ok {
		lock := val.(*EntityLock)
		lock.mu.RLock()
		isOwner := lock.SessionID == sessionID
		lock.mu.RUnlock()

		if isOwner {
			entityLockTable.Delete(entityID)
			return true
		}
	}
	return false
}

// UnlockAllByClientID releases all locks owned by a specific client
func UnlockAllByClientID(clientID string) int {
	count := 0
	entityLockTable.Range(func(key, value interface{}) bool {
		lock := value.(*EntityLock)
		lock.mu.RLock()
		isOwner := lock.SessionID == clientID
		lock.mu.RUnlock()

		if isOwner {
			entityLockTable.Delete(key)
			count++
		}
		return true
	})
	return count
}

// GetLockInfo retrieves information about a lock
func GetLockInfo(entityID int64) (*EntityLock, bool) {
	if val, ok := entityLockTable.Load(entityID); ok {
		lock := val.(*EntityLock)
		lock.mu.RLock()
		defer lock.mu.RUnlock()

		// Return a copy to avoid race conditions
		return &EntityLock{
			EntityID:   lock.EntityID,
			EntityType: lock.EntityType,
			Username:   lock.Username,
			SessionID:  lock.SessionID,
			LockedAt:   lock.LockedAt,
			LastAccess: lock.LastAccess,
		}, true
	}
	return nil, false
}

// CleanupStaleLocks removes locks with no activity for the specified duration
func CleanupStaleLocks(maxIdle time.Duration) int {
	count := 0
	now := time.Now()

	entityLockTable.Range(func(key, value interface{}) bool {
		lock := value.(*EntityLock)
		lock.mu.RLock()
		isStale := now.Sub(lock.LastAccess) > maxIdle
		lock.mu.RUnlock()

		if isStale {
			entityLockTable.Delete(key)
			count++
		}
		return true
	})

	return count
}

// GetAllLocks returns all active locks (for monitoring/debugging)
func GetAllLocks() []EntityLock {
	var locks []EntityLock

	entityLockTable.Range(func(key, value interface{}) bool {
		lock := value.(*EntityLock)
		lock.mu.RLock()
		locks = append(locks, EntityLock{
			EntityID:   lock.EntityID,
			EntityType: lock.EntityType,
			Username:   lock.Username,
			SessionID:  lock.SessionID,
			LockedAt:   lock.LockedAt,
			LastAccess: lock.LastAccess,
		})
		lock.mu.RUnlock()
		return true
	})

	return locks
}

// FormatLockError creates a user-friendly error message for a locked resource
func FormatLockError(lock *EntityLock) string {
	duration := time.Since(lock.LockedAt)
	minutes := int(duration.Minutes())

	entityName := lock.EntityType
	if entityName == "" {
		entityName = "Zapis"
	}

	if minutes < 1 {
		return fmt.Sprintf("%s je zaključan od strane korisnika '%s' (upravo sada)", entityName, lock.Username)
	} else if minutes == 1 {
		return fmt.Sprintf("%s je zaključan od strane korisnika '%s' (pre 1 minut)", entityName, lock.Username, minutes)
	} else if minutes < 60 {
		return fmt.Sprintf("%s je zaključan od strane korisnika '%s' (pre %d minuta)", entityName, lock.Username, minutes)
	} else {
		hours := minutes / 60
		return fmt.Sprintf("%s je zaključan od strane korisnika '%s' (pre %d sati)", entityName, lock.Username, hours)
	}
}

// Legacy wrapper functions for backward compatibility
func TryLockNalog(nalogID int64, username, sessionID string) (bool, *EntityLock, error) {
	return TryLockEntity(nalogID, "Nalog", username, sessionID)
}

func UnlockNalog(nalogID int64, sessionID string) bool {
	return UnlockEntity(nalogID, sessionID)
}
