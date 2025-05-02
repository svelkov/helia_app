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

// Secret key for signing the JWT (keep this secure!)
var jwtSecret = []byte("3285f0d71eed0c41fded2115c9cc8ac09a0ab5a519565df10afdb20f8013c5268f2c19948b6af096c1cfc2921ab086be21fa5407b9d91aeb08eeeeef3c2e16c9ae30ae15f27d340f17c450468fef50795e58bb7351a94602bc045aea1a1ff3b03039081208cf067b44fd913b98b712e34ba080941f5ff8545b0eac26824f0ef4a93109939d8f917e1fac1eb588f4272ebac415975bcdc994c3a0fea7c3805d601443ad71dd9043858de5c2bfe64106683d9eaebce28442ce7bb22298d5b85cc3cc41e6f81f9c0f8f678cce559f745645edc5a5009ba20f8b5a16be4ee7dada7791913c90e3629a44b88a17d3d107bd3a6c0f3000b4865b2c015c0875901a028e")

func NewJWTAdapter(secretKey []byte) *JWTAdapter {
	return &JWTAdapter{secretKey: secretKey}
}

// GenerateJWT generates a JWT for the given username.
func GenerateJWT(username string) (string, error) {
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
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyJWT verifies a JWT and returns the claims.
func VerifyJWT(tokenString string) (*domain.UserClaims, error) {
	claims := &domain.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
