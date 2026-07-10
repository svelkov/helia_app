package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"helia/internal/common"
	"helia/internal/infrastructure/db"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// EntityLock represents a lock record in the database
type EntityLock struct {
	ID         int64     `db:"id"`
	EntityType string    `db:"entitytype"`
	EntityID   int64     `db:"entityid"`
	LockedBy   string    `db:"lockedby"`
	LockedAt   time.Time `db:"lockedat"`
	ExpiresAt  time.Time `db:"expiresat"`
}

// LockService handles entity locking and unlocking
type LockService struct {
	db db.Database
}

// NewLockService creates a new lock service
func NewLockService(dbInst db.Database) *LockService {
	return &LockService{db: dbInst}
}

// Lock attempts to acquire a lock on an entity
// Returns error if lock is already held by another user
func (s *LockService) Lock(ctx context.Context, entityType string, entityId int64, userId string) error {
	expiresAt := time.Now().Add(10 * time.Minute)
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Try to insert a new lock record
	qb := common.NewQueryBuilder(`INSERT INTO entitylocks 
	(entitytype, entityid, lockedby, expiresat) VALUES ($1, $2, $3, $4)
	ON CONFLICT (entitytype, entityid) DO NOTHING`, false)

	qb.AddArgs(entityType, entityId, userId, expiresAt)

	sqlQuery, args := qb.Build()
	result, err := tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}

	if rows == 0 {
		// Lock already exists, check who holds it
		var lockedBy string
		var expiresBy time.Time
		err := tx.QueryRowContext(ctx,
			`SELECT lockedby, expiresat
			 FROM entitylocks
			 WHERE entitytype = $1 AND entityid = $2
			 FOR UPDATE`,
			entityType, entityId,
		).Scan(&lockedBy, &expiresBy)

		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("query error: %w", err)
		}
		if err == sql.ErrNoRows {
			return fmt.Errorf("entity lock state changed, please retry")
		}

		// If lock is expired, try to delete and re-lock
		if expiresBy.Before(time.Now()) {
			_, err := tx.ExecContext(ctx,
				`DELETE FROM entitylocks
				 WHERE entitytype = $1 AND entityid = $2 AND expiresat < NOW()`,
				entityType, entityId,
			)
			if err != nil {
				return fmt.Errorf("delete expired lock error: %w", err)
			}

			_, err = tx.ExecContext(ctx,
				`INSERT INTO entitylocks (entitytype, entityid, lockedby, expiresat)
				 VALUES ($1, $2, $3, $4)
				 ON CONFLICT (entitytype, entityid) DO NOTHING`,
				entityType, entityId, userId, expiresAt,
			)
			if err != nil {
				return fmt.Errorf("re-lock error: %w", err)
			}

			return tx.Commit()
		}

		if lockedBy == userId {
			// User owns this lock, refresh expiry
			_, err := tx.ExecContext(ctx,
				`UPDATE entitylocks
				 SET expiresat = $1
				 WHERE entitytype = $2 AND entityid = $3 AND lockedby = $4`,
				expiresAt, entityType, entityId, userId,
			)
			if err != nil {
				return err
			}
			return tx.Commit()
		}

		return fmt.Errorf("entity locked by another user")
	}

	return tx.Commit()
}

// VerifyLock checks if the current user still holds the lock
// Returns error if lock is not held or has expired
func (s *LockService) VerifyLock(ctx context.Context, entityType string, entityId int64, userId string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	var lockedBy string
	var expiresAt time.Time

	qb := common.NewQueryBuilder(`SELECT lockedby, expiresat FROM entitylocks`, true)
	qb.AddEqual("entitytype", entityType)
	qb.AddEqual("entityid", entityId)

	sqlQuery, args := qb.Build()
	err = tx.QueryRowContext(ctx, sqlQuery, args...).Scan(&lockedBy, &expiresAt)
	if err == sql.ErrNoRows {
		log.Printf("no active lock on this entity")
		return nil
	}

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Check if lock has expired
	if expiresAt.Before(time.Now()) {
		qb := common.NewQueryBuilder(`DELETE FROM entitylocks `, true)
		qb.AddEqual("entitytype", entityType)
		qb.AddEqual("entityid", entityId)

		sqlQuery, args := qb.Build()
		_, err := tx.ExecContext(ctx, sqlQuery, args...)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
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
// Only the user who holds the lock can release it
func (s *LockService) Unlock(ctx context.Context, entityType string, entityId int64, userId string) error {
	qb := common.NewQueryBuilder(`DELETE FROM entitylocks`, true)
	qb.AddEqual("entitytype", entityType)
	qb.AddEqual("entityid", entityId)
	qb.AddEqual("lockedby", userId)

	sqlQuery, args := qb.Build()
	result, err := s.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}
	if rows == 0 {
		// No lock to release (could be already unlocked or held by someone else)
		return nil // Silent success - idempotent
	}

	return nil
}

// CleanupExpired removes all expired locks from the database
// Should be called periodically (e.g., every 5 minutes)
func (s *LockService) CleanupExpired(ctx context.Context) error {
	qb := common.NewQueryBuilder(`DELETE FROM entitylocks WHERE expiresat < NOW()`, false)
	sqlQuery, args := qb.Build()
	_, err := s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}

// GetLockInfo returns information about a lock, if it exists
func (s *LockService) GetLockInfo(ctx context.Context, entityType string, entityId int64) (*EntityLock, error) {
	lock := &EntityLock{}
	qb := common.NewQueryBuilder(`SELECT id, entitytype, entityid, lockedby, lockedat, expiresat
		 FROM entitylocks`, false)
	qb.AddEqual("entitytype", entityType)
	qb.AddEqual("entityid", entityId)
	qb.AddCondition("expiresat", time.Now(), ">")

	qb.AddArgs(entityType, entityId)
	sqlQuery, args := qb.Build()
	err := s.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&lock.ID, &lock.EntityType, &lock.EntityID, &lock.LockedBy, &lock.LockedAt, &lock.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return lock, nil
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
