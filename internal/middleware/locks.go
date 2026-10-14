package middleware

import (
	"context"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// EntityLock represents a lock record in the database
type EntityLock = domain.EntityLock

// LockService handles entity locking and unlocking; the SQL is in repository.LockRepository.
type LockService struct {
	repo *repository.LockRepository
}

// NewLockService creates a new lock service
func NewLockService(dbInst db.Database) *LockService {
	return &LockService{repo: repository.NewLockRepository(dbInst)}
}

// Lock attempts to acquire a lock on an entity
// Returns error if lock is already held by another user
func (s *LockService) Lock(ctx context.Context, entityType string, entityId int64, userId string) error {
	expiresAt := time.Now().Add(10 * time.Minute)
	tx, err := s.repo.BeginTx()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Try to insert a new lock record
	inserted, err := s.repo.InsertTx(ctx, tx, entityType, entityId, userId, expiresAt)
	if err != nil {
		return err
	}
	if inserted {
		return tx.Commit()
	}

	// Lock already exists, check who holds it
	lockedBy, expiresBy, found, err := s.repo.HolderTx(ctx, tx, entityType, entityId, true)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("entity lock state changed, please retry")
	}

	// If lock is expired, try to delete and re-lock
	if expiresBy.Before(time.Now()) {
		if err := s.repo.DeleteTx(ctx, tx, entityType, entityId, true); err != nil {
			return err
		}
		if _, err := s.repo.InsertTx(ctx, tx, entityType, entityId, userId, expiresAt); err != nil {
			return err
		}
		return tx.Commit()
	}

	if lockedBy == userId {
		// User owns this lock, refresh expiry
		if err := s.repo.RefreshTx(ctx, tx, entityType, entityId, userId, expiresAt); err != nil {
			return err
		}
		return tx.Commit()
	}

	return fmt.Errorf("entity locked by another user")
}

// VerifyLock checks if the current user still holds the lock
// Returns error if lock is not held or has expired
func (s *LockService) VerifyLock(ctx context.Context, entityType string, entityId int64, userId string) error {
	tx, err := s.repo.BeginTx()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	lockedBy, expiresAt, found, err := s.repo.HolderTx(ctx, tx, entityType, entityId, false)
	if err != nil {
		return err
	}
	if !found {
		log.Printf("no active lock on this entity")
		return nil
	}

	// Check if lock has expired
	if expiresAt.Before(time.Now()) {
		if err := s.repo.DeleteTx(ctx, tx, entityType, entityId, false); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit error: %w", err)
		}
		log.Printf("lock expired and removed")
		return nil
	}

	// Check if lock is held by this user
	if lockedBy != userId {
		return fmt.Errorf("slog je zaključan od strane drugog korisnika")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}

	return nil
}

// Unlock releases a lock on an entity
// Only the user who holds the lock can release it; releasing a lock that is not there (already
// unlocked or held by someone else) is a silent success.
func (s *LockService) Unlock(ctx context.Context, entityType string, entityId int64, userId string) error {
	return s.repo.Release(ctx, entityType, entityId, userId)
}

// CleanupExpired removes all expired locks from the database
// Should be called periodically (e.g., every 5 minutes)
func (s *LockService) CleanupExpired(ctx context.Context) error {
	return s.repo.DeleteExpired(ctx)
}

// GetLockInfo returns information about a lock, if it exists
func (s *LockService) GetLockInfo(ctx context.Context, entityType string, entityId int64) (*EntityLock, error) {
	return s.repo.Active(ctx, entityType, entityId, time.Now())
}

// LockMiddleware wraps lock operations for Gin handlers
type LockMiddleware struct {
	lockService *LockService
}

// NewLockMiddleware creates a new lock middleware
func NewLockMiddleware(lockService *LockService) *LockMiddleware {
	return &LockMiddleware{lockService: lockService}
}

// WithEntityLock is a middleware that acquires a lock before the handler runs
// and automatically releases it when the handler completes
// entityType: type of entity being locked (e.g., "fnal", "partneri")
// paramName: name of the URL parameter containing the entity ID (e.g., "id")
func (lm *LockMiddleware) WithEntityLock(entityType, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract entity ID from URL parameter
		entityIdStr := c.Param(paramName)
		entityId, err := strconv.ParseInt(entityIdStr, 10, 64)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
			c.Abort()
			return
		}

		// Get user ID from context (set by auth middleware)
		userId, exists := c.Get("userid")
		if !exists {
			common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
			c.Abort()
			return
		}

		userIdStr := userId.(string)

		// Attempt to acquire lock
		lockErr := lm.lockService.Lock(c.Request.Context(), entityType, entityId, userIdStr)
		if lockErr != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgLockFailed)
			c.Abort()
			return
		}

		// Store lock information in context for handler access
		c.Set("entitytype", entityType)
		c.Set("entityid", entityId)
		c.Set("lockacquired", true)

		// Defer unlock - ensures lock is released even if handler panics
		defer func() {
			if err := lm.lockService.Unlock(c.Request.Context(), entityType, entityId, userIdStr); err != nil {
				// Log error but don't fail - unlock is best effort
				fmt.Printf("Error unlocking %s:%d: %v\n", entityType, entityId, err)
			}
		}()

		// Continue to next handler
		c.Next()
	}
}

// WithEntityLockHold acquires a lock but DOES NOT release it when request completes
// Use this for the FIRST request in a multi-request operation (e.g., FNAL update)
// The lock will be held for 10 minutes and must be manually released or will expire
// entityType: type of entity being locked (e.g., "fnal", "partneri")
// paramName: name of the URL parameter containing the entity ID (e.g., "id")
func (lm *LockMiddleware) WithEntityLockHold(entityType, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract entity ID from URL parameter
		entityIdStr := c.Param(paramName)
		if entityIdStr == "" {
			entityIdStr = c.Query(paramName) // Try query parameter if not in path
		}
		entityId, err := strconv.ParseInt(entityIdStr, 10, 64)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
			c.Abort()
			return
		}

		// Get user ID from context (set by auth middleware)
		userId, exists := c.Get("username")
		if !exists {
			common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
			c.Abort()
			return
		}

		userIdStr := userId.(string)

		// Attempt to acquire lock
		lockErr := lm.lockService.Lock(c.Request.Context(), entityType, entityId, userIdStr)
		if lockErr != nil {
			common.WriteJSONResponse(c, http.StatusConflict, false, nil, common.ErrMsgLockFailed)
			c.Abort()
			return
		}

		// Store lock information in context for handler access
		c.Set("entitytype", entityType)
		c.Set("entityid", entityId)
		c.Set("lockacquired", true)

		// NO DEFER UNLOCK - lock is held for subsequent requests
		// Continue to next handler
		c.Next()
	}
}

// WithEntityLockVerifyAndRefresh verifies an existing lock AND refreshes its expiry
// Use this for INTERMEDIATE requests that need to maintain the lock (e.g., FPRO insert)
// The lock expiry is extended by 10 minutes and NOT released when request completes
// entityType: type of entity being locked (e.g., "fnal")
// paramName: name of the URL parameter containing the entity ID (e.g., "fnal_id")
func (lm *LockMiddleware) WithEntityLockVerifyAndRefresh(entityType, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract entity ID from URL parameter
		entityIdStr := c.Param(paramName)
		entityId, err := strconv.ParseInt(entityIdStr, 10, 64)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
			c.Abort()
			return
		}

		// Get user ID from context (set by auth middleware)
		userId, exists := c.Get("username")
		if !exists {
			common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
			c.Abort()
			return
		}

		userIdStr := userId.(string)

		// Verify the lock is still held by this user
		verifyErr := lm.lockService.VerifyLock(c.Request.Context(), entityType, entityId, userIdStr)
		if verifyErr != nil {
			common.WriteJSONResponse(c, http.StatusConflict, false, nil, common.ErrMsgLockFailed)
			c.Abort()
			return
		}

		// Refresh the lock (extend expiry by 10 minutes)
		refreshErr := lm.lockService.Lock(c.Request.Context(), entityType, entityId, userIdStr)
		if refreshErr != nil {
			common.WriteJSONResponse(c, http.StatusConflict, false, nil, common.ErrMsgLockFailed)
			c.Abort()
			return
		}

		// Store lock information in context for handler access
		c.Set("entitytype", entityType)
		c.Set("entityid", entityId)
		c.Set("lockverified", true)

		// NO DEFER UNLOCK - lock continues to be held
		// Continue to next handler
		c.Next()
	}
}

// WithEntityLockVerifyAndRelease verifies an existing lock AND releases it when request completes
// Use this for the FINAL request in a multi-request operation (e.g., FNAL commit)
// entityType: type of entity being locked (e.g., "fnal")
// paramName: name of the URL parameter containing the entity ID (e.g., "id")
func (lm *LockMiddleware) WithEntityLockVerifyAndRelease(entityType, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract entity ID from URL parameter
		entityIdStr := c.Param(paramName)
		entityId, err := strconv.ParseInt(entityIdStr, 10, 64)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
			c.Abort()
			return
		}

		// Get user ID from context (set by auth middleware)
		userId, exists := c.Get("username")
		if !exists {
			common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
			c.Abort()
			return
		}

		userIdStr := userId.(string)

		// Verify the lock is still held by this user
		verifyErr := lm.lockService.VerifyLock(c.Request.Context(), entityType, entityId, userIdStr)
		if verifyErr != nil {
			common.WriteJSONResponse(c, http.StatusConflict, false, nil, common.ErrMsgLockFailed)
			c.Abort()
			return
		}

		// Store lock information in context for handler access
		c.Set("entitytype", entityType)
		c.Set("entityid", entityId)
		c.Set("lockverified", true)

		// Defer unlock - release lock when this final request completes
		defer func() {
			if err := lm.lockService.Unlock(c.Request.Context(), entityType, entityId, userIdStr); err != nil {
				fmt.Printf("Error unlocking %s:%d: %v\n", entityType, entityId, err)
			}
		}()

		// Continue to next handler
		c.Next()
	}
}

// VerifyLock checks if the lock is still held before operation
// Use this in handlers that need extra safety (when not using middleware)
func (lm *LockMiddleware) VerifyLock(c *gin.Context, entityType string, entityId int64) error {
	userId, exists := c.Get("username")
	if !exists {
		return fmt.Errorf("unauthorized")
	}

	return lm.lockService.VerifyLock(c.Request.Context(), entityType, entityId, userId.(string))
}
