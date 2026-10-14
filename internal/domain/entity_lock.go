package domain

import "time"

// EntityLock is a lock record of the entitylocks table: an entity (type + id) locked by a user
// until expiresat. Locks have no tenant: the entity id is unique on its own.
type EntityLock struct {
	ID         int64     `db:"id"`
	EntityType string    `db:"entitytype"`
	EntityID   int64     `db:"entityid"`
	LockedBy   string    `db:"lockedby"`
	LockedAt   time.Time `db:"lockedat"`
	ExpiresAt  time.Time `db:"expiresat"`
}
