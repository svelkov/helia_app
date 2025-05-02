package service

import "helia/internal/domain"

// UserService defines the interface for user-related operations, like user retrieval.
type UserService interface {
	GetUserByUsername(username string) (*domain.User, error)
}
