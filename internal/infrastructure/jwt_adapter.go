package infrastructure

import (
	"fmt"
	"helia/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAdapter struct {
	secretKey []byte
}

func NewJWTAdapter(secretKey []byte) *JWTAdapter {
	return &JWTAdapter{secretKey: secretKey}
}

// GenerateJWT generates a JWT for the given username.
func GenerateJWT(username string, secretKey []byte) (string, error) {
	claims := domain.UserClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "HELIA",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyJWT verifies a JWT and returns the claims.
func VerifyJWT(tokenString string, secretKey []byte) (*domain.UserClaims, error) {
	claims := &domain.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
