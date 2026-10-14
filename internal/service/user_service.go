package service

import (
	"context"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
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
	// Query to find user by username. Where (not AddEqual): an empty username must find nobody.
	b := querybuilder.New(`SELECT id, username, email, name, password_hash, two_fa_enabled, two_fa_secret, backup_codes
	         FROM appusers`).
		Where("username = ?", username).
		Limit(1)
	users, err := s.userRepo.Select(ctx, b)
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
	// One INSERT ... RETURNING (atomic on its own; the old explicit transaction around it added nothing).
	_, err := repository.InsertUser(ctx, s.userRepo, user)
	return err
}

// UpdateUser updates an existing user
func (s *userResource) UpdateUser(ctx context.Context, user *domain.User) error {
	// repository.ErrNotFound when the id matches no user.
	return repository.UpdateUser(ctx, s.userRepo, user)
}

// UserExists checks if a user exists by username
func (s *userResource) UserExists(ctx context.Context, username string) (bool, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}
