package service

import (
	"context"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/validation"
)

// UserService defines the interface for user-related operations
type UserService interface {
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
	UserExists(ctx context.Context, username string) (bool, error)
}

// userServiceImpl implements UserService using database access
type userResource struct {
	userRepo  *repository.BaseRepository[domain.User]
	validator validation.Validator[domain.User]
}

// NewUserService creates a new instance of UserService
func NewUserService(userRepo *repository.BaseRepository[domain.User], validator validation.Validator[domain.User]) UserService {
	return &userResource{
		userRepo:  userRepo,
		validator: validator,
	}
}

// GetUserByUsername retrieves a user by username using custom query
func (s *userResource) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	// Query to find user by username
	qb := common.NewQueryBuilder(`SELECT id, username, email, name, password_hash, two_fa_enabled, two_fa_secret, backup_codes 
	         FROM appusers`, true)
	qb.AddEqual("username", username)

	sqlQuery, args := qb.Build()
	users, err := s.userRepo.GetAllCustom(ctx, sqlQuery, "", args, " LIMIT 1 ", "")
	if err != nil {
		return nil, err
	}

	if users == nil || len(*users) == 0 {
		return nil, nil // User not found, return nil without error
	}

	return &(*users)[0], nil
}

// CreateUser saves a new user to the database using raw SQL
func (s *userResource) CreateUser(ctx context.Context, user *domain.User) error {
	// Use raw SQL insert via the database abstraction
	userID := int64(0)
	qb := common.NewQueryBuilder(`INSERT INTO appusers (username, email, name, password_hash, two_fa_enabled, two_fa_secret, backup_codes) 
	         VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, false)
	tx, err := s.userRepo.BeginTx()
	if err != nil {
		return err
	}
	sqlQuery, args := qb.Build()
	args = append(args, user.Username, user.Email, user.Name, user.PasswordHash, user.TwoFAEnabled, user.TwoFASecret, user.BackupCodes)
	// Execute the insert and get the generated ID
	err = tx.QueryRowContext(ctx, sqlQuery, args...).Scan(&userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateUser updates an existing user
func (s *userResource) UpdateUser(ctx context.Context, user *domain.User) error {
	qb := common.NewQueryBuilder(`UPDATE appusers 
	         SET username = $1, email = $2, name = $3, password_hash = $4, 
	             two_fa_enabled = $5, two_fa_secret = $6, backup_codes = $7`, false)

	qb.AddArgs(user.Username, user.Email, user.Name, user.PasswordHash, user.TwoFAEnabled, user.TwoFASecret, user.BackupCodes, user.Id)
	qb.AddEqual("id", user.Id)
	sqlQuery, args := qb.Build()

	tx, err := s.userRepo.BeginTx()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// UserExists checks if a user exists by username
func (s *userResource) UserExists(ctx context.Context, username string) (bool, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}
