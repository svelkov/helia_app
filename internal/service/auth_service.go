package service

import (
	"fmt"
	"helia/internal/domain"
)

const ErrInvalidCredentials = "Netacno korisnicko ime ili password..."

// AuthService defines the interface for authentication-related operations.
type AuthService interface {
	Login(username, password string) (string, string, error) // Return access & refresh tokens
	Logout(refreshToken string) error
	RefreshToken(refreshToken string) (string, error)
}

// JWTService defines the interface for JWT-related operations.
type JWTService interface {
	GenerateAccessToken(username string) (string, error)
	GenerateRefreshToken(username string) (string, error)
	VerifyAccessToken(tokenString string) (*domain.UserClaims, error)
	VerifyRefreshToken(tokenString string) (string, error) //return username
	InvalidateRefreshToken(tokenString string) error
}

// authServiceImpl implements the AuthService interface.
type authServiceImpl struct {
	jwtService          JWTService
	userService         UserService
	refreshTokenStorage map[string]string //Simulate refresh token storage
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(jwtService JWTService, userService UserService) AuthService {
	return &authServiceImpl{
		jwtService:          jwtService,
		userService:         userService,
		refreshTokenStorage: make(map[string]string),
	}
}

// Login handles user login and generates a JWT.
func (s *authServiceImpl) Login(username, password string) (string, string, error) {
	// user, err := s.userService.GetUserByUsername(username)
	// if err != nil {
	// 	return "", "", err // User not found or error retrieving user.
	// }

	// if !user.VerifyPassword(password) { // Assuming User has a VerifyPassword method
	// 	return "", "", errors.New(ErrInvalidCredentials)
	// }

	accessToken, err := s.jwtService.GenerateAccessToken(username)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(username)
	if err != nil {
		return "", "", err
	}

	s.refreshTokenStorage[refreshToken] = username //Store refresh token
	return accessToken, refreshToken, nil
}

func (s *authServiceImpl) Logout(refreshToken string) error {
	delete(s.refreshTokenStorage, refreshToken)
	return s.jwtService.InvalidateRefreshToken(refreshToken)
}

func (s *authServiceImpl) RefreshToken(refreshToken string) (string, error) {
	username, err := s.jwtService.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	if storedUsername, ok := s.refreshTokenStorage[refreshToken]; !ok || storedUsername != username {
		return "", fmt.Errorf("invalid refresh token")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(username)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
