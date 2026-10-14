package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// LockRepository reads and writes the entity locks (table entitylocks). A lock has no tenant: the
// entity is identified by its type and id, the owner by lockedby.
type LockRepository struct {
	db db.Database
}

// NewLockRepository returns the repository of the entity locks.
func NewLockRepository(database db.Database) *LockRepository {
	return &LockRepository{db: database}
}

// BeginTx starts a transaction for a lock operation.
func (r *LockRepository) BeginTx() (db.Transaction, error) {
	return r.db.Beginx()
}

func lockWhere(b *querybuilder.Builder, entityType string, entityID int64) *querybuilder.Builder {
	return b.Where("entitytype = ?", entityType).Where("entityid = ?", entityID)
}

// InsertTx inserts the lock unless the entity is locked already; it reports whether a row was
// inserted.
func (r *LockRepository) InsertTx(ctx context.Context, tx db.Transaction, entityType string, entityID int64, user string, expiresAt time.Time) (bool, error) {
	q, args, err := querybuilder.New(`INSERT INTO entitylocks (entitytype, entityid, lockedby, expiresat) VALUES (?, ?, ?, ?)
		ON CONFLICT (entitytype, entityid) DO NOTHING`, entityType, entityID, user, expiresAt).Build()
	if err != nil {
		return false, fmt.Errorf("build lock insert: %w", err)
	}
	res, err := tx.ExecContext(ctx, q, args...)
	if err != nil {
		return false, fmt.Errorf("insert lock %s %d: %w", entityType, entityID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("insert lock %s %d: %w", entityType, entityID, err)
	}
	return n > 0, nil
}

// HolderTx returns the owner and the expiry of the lock of an entity; found is false when the
// entity is not locked. forUpdate locks the row for the rest of the transaction.
func (r *LockRepository) HolderTx(ctx context.Context, tx db.Transaction, entityType string, entityID int64, forUpdate bool) (lockedBy string, expiresAt time.Time, found bool, err error) {
	b := lockWhere(querybuilder.New(`SELECT lockedby, expiresat FROM entitylocks`), entityType, entityID)
	if forUpdate {
		b.ForUpdate()
	}
	q, args, err := b.Build()
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("build lock select: %w", err)
	}
	err = tx.QueryRowContext(ctx, q, args...).Scan(&lockedBy, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", time.Time{}, false, nil
	}
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("read lock %s %d: %w", entityType, entityID, err)
	}
	return lockedBy, expiresAt, true, nil
}

// DeleteTx removes the lock of an entity, whoever holds it. With onlyExpired only an expired lock
// is removed.
func (r *LockRepository) DeleteTx(ctx context.Context, tx db.Transaction, entityType string, entityID int64, onlyExpired bool) error {
	b := lockWhere(querybuilder.New(`DELETE FROM entitylocks`), entityType, entityID)
	if onlyExpired {
		b.Where("expiresat < NOW()")
	}
	q, args, err := b.Build()
	if err != nil {
		return fmt.Errorf("build lock delete: %w", err)
	}
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("delete lock %s %d: %w", entityType, entityID, err)
	}
	return nil
}

// RefreshTx moves the expiry of a lock the user holds.
func (r *LockRepository) RefreshTx(ctx context.Context, tx db.Transaction, entityType string, entityID int64, user string, expiresAt time.Time) error {
	b := lockWhere(querybuilder.NewUpdate("entitylocks").Set("expiresat = ?", expiresAt), entityType, entityID).
		Where("lockedby = ?", user)
	q, args, err := b.Build()
	if err != nil {
		return fmt.Errorf("build lock refresh: %w", err)
	}
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("refresh lock %s %d: %w", entityType, entityID, err)
	}
	return nil
}

// Release removes the lock of an entity held by the user. lockedby is always a condition, so an
// empty user can not release another user's lock. Releasing a lock that is not there is not an
// error.
func (r *LockRepository) Release(ctx context.Context, entityType string, entityID int64, user string) error {
	q, args, err := lockWhere(querybuilder.New(`DELETE FROM entitylocks`), entityType, entityID).
		Where("lockedby = ?", user).
		Build()
	if err != nil {
		return fmt.Errorf("build unlock: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("unlock %s %d: %w", entityType, entityID, err)
	}
	return nil
}

// DeleteExpired removes every expired lock.
func (r *LockRepository) DeleteExpired(ctx context.Context) error {
	q, args, err := querybuilder.New(`DELETE FROM entitylocks`).Where("expiresat < NOW()").Build()
	if err != nil {
		return fmt.Errorf("build expired locks cleanup: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("delete expired locks: %w", err)
	}
	return nil
}

// Active returns the lock of an entity that has not expired at now, or nil. (The key of
// entitylocks is entitylocksid; the old query read a column id that does not exist.)
func (r *LockRepository) Active(ctx context.Context, entityType string, entityID int64, now time.Time) (*domain.EntityLock, error) {
	q, args, err := lockWhere(querybuilder.New(`SELECT entitylocksid AS id, entitytype, entityid, lockedby, lockedat, expiresat FROM entitylocks`), entityType, entityID).
		Where("expiresat > ?", now).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build lock info: %w", err)
	}
	lock := &domain.EntityLock{}
	err = r.db.QueryRowContext(ctx, q, args...).Scan(&lock.ID, &lock.EntityType, &lock.EntityID, &lock.LockedBy, &lock.LockedAt, &lock.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read lock %s %d: %w", entityType, entityID, err)
	}
	return lock, nil
}
