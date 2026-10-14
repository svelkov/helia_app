package repository

import (
	"context"
	"fmt"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
)

// InsertUser inserts an application user and returns its id. appusers has no tenant: a user
// chooses the firm and the period after the login.
func InsertUser(ctx context.Context, r *BaseRepository[domain.User], u *domain.User) (int64, error) {
	q, args, err := querybuilder.New(`INSERT INTO appusers (username, email, name, password_hash, two_fa_enabled, two_fa_secret, backup_codes)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		u.Username, u.Email, u.Name, u.PasswordHash, u.TwoFAEnabled, u.TwoFASecret, u.BackupCodes).Build()
	if err != nil {
		return 0, fmt.Errorf("build user insert: %w", err)
	}
	var id int64
	if err := r.DB.QueryRowContext(ctx, q, args...).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert user %q: %w", u.Username, err)
	}
	return id, nil
}

// UpdateUser writes the login data of an application user; ErrNotFound when the id matches no
// user.
func UpdateUser(ctx context.Context, r *BaseRepository[domain.User], u *domain.User) error {
	q, args, err := querybuilder.NewUpdate("appusers").
		Set("username = ?, email = ?, name = ?, password_hash = ?", u.Username, u.Email, u.Name, u.PasswordHash).
		Set("two_fa_enabled = ?, two_fa_secret = ?, backup_codes = ?", u.TwoFAEnabled, u.TwoFASecret, u.BackupCodes).
		Where("id = ?", u.Id).
		Build()
	if err != nil {
		return fmt.Errorf("build user update: %w", err)
	}
	res, err := r.DB.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("update user %d: %w", u.Id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user %d: %w", u.Id, err)
	}
	if n == 0 {
		return fmt.Errorf("update user %d: %w", u.Id, ErrNotFound)
	}
	return nil
}
