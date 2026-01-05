package infrastructure

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"helia/internal/domain"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenConfig holds token configuration
type TokenConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CSRFTokenTTL    time.Duration
	EnableBlacklist bool
	CleanupInterval time.Duration // How often to cleanup expired tokens
}

// TokenManager handles all token operations with JWT-only authentication
type TokenManager struct {
	secretKey      []byte
	config         TokenConfig
	tokenBlacklist map[string]time.Time // jti -> revocation time
	blacklistMutex sync.RWMutex
	refreshTokenDB map[string]*RefreshTokenData // In-memory, should be replaced with DB
	refreshDBMutex sync.RWMutex
}

// RefreshTokenData stores refresh token metadata
type RefreshTokenData struct {
	UserID    string
	Token     string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// NewTokenManager creates a new token manager with configuration
func NewTokenManager(secretKey []byte, config TokenConfig) *TokenManager {
	tm := &TokenManager{
		secretKey:      secretKey,
		config:         config,
		tokenBlacklist: make(map[string]time.Time),
		refreshTokenDB: make(map[string]*RefreshTokenData),
	}

	// Start cleanup goroutine if enabled
	if config.EnableBlacklist && config.CleanupInterval > 0 {
		go tm.cleanupExpiredTokens()
	}

	return tm
}

// GenerateTokenPair generates both access and refresh tokens
func (tm *TokenManager) GenerateTokenPair(username string) (*TokenPair, error) {
	csrfToken, err := tm.GenerateCSRFToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	accessToken, err := tm.GenerateAccessToken(username, csrfToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := tm.GenerateRefreshToken(username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CSRFToken:    csrfToken,
		ExpiresIn:    int(tm.config.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// GenerateAccessToken generates a short-lived access token with CSRF token hash
func (tm *TokenManager) GenerateAccessToken(username, csrfToken string) (string, error) {
	now := time.Now()
	csrfHash := hashToken(csrfToken)

	claims := &domain.UserClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "HELIA",
			ID:        generateJTI(),
			Subject:   username,
		},
	}

	// Store CSRF hash in claims for validation
	claims.CSRFHash = csrfHash
	claims.TokenType = "access"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a long-lived refresh token
func (tm *TokenManager) GenerateRefreshToken(username string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(tm.config.RefreshTokenTTL)
	jti := generateJTI()

	claims := &domain.UserClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "HELIA",
			ID:        jti,
			Subject:   username,
		},
	}

	claims.TokenType = "refresh"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store in memory (should be in database for production)
	tm.refreshDBMutex.Lock()
	tm.refreshTokenDB[jti] = &RefreshTokenData{
		UserID:    username,
		Token:     tokenString,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}
	tm.refreshDBMutex.Unlock()

	return tokenString, nil
}

// GenerateCSRFToken generates a standalone CSRF token
func (tm *TokenManager) GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// VerifyAccessToken verifies and returns claims from access token
func (tm *TokenManager) VerifyAccessToken(tokenString string) (*domain.UserClaims, error) {
	claims := &domain.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check blacklist
	if tm.config.EnableBlacklist && claims.ID != "" {
		tm.blacklistMutex.RLock()
		_, isBlacklisted := tm.tokenBlacklist[claims.ID]
		tm.blacklistMutex.RUnlock()

		if isBlacklisted {
			return nil, fmt.Errorf("token has been revoked")
		}
	}

	return claims, nil
}

// VerifyRefreshToken verifies and returns claims from refresh token
func (tm *TokenManager) VerifyRefreshToken(tokenString string) (*domain.UserClaims, error) {
	claims := &domain.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Verify it's actually a refresh token
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("not a refresh token")
	}

	// Check in database
	if claims.ID != "" {
		tm.refreshDBMutex.RLock()
		tokenData, exists := tm.refreshTokenDB[claims.ID]
		tm.refreshDBMutex.RUnlock()

		if !exists || tokenData.Token != tokenString {
			return nil, fmt.Errorf("refresh token not found or invalid")
		}
	}

	return claims, nil
}

// VerifyCSRFToken verifies CSRF token against the hash in JWT claims
func (tm *TokenManager) VerifyCSRFToken(csrfToken string, accessToken string) (bool, error) {
	if csrfToken == "" || accessToken == "" {
		return false, fmt.Errorf("missing CSRF or access token")
	}

	claims, err := tm.VerifyAccessToken(accessToken)
	if err != nil {
		return false, err
	}

	// Compare CSRF hash from token claims
	calculatedHash := hashToken(csrfToken)
	return calculatedHash == claims.CSRFHash, nil
}

// RevokeAccessToken adds token to blacklist (for logout)
func (tm *TokenManager) RevokeAccessToken(tokenID string) {
	if !tm.config.EnableBlacklist {
		return
	}

	tm.blacklistMutex.Lock()
	tm.tokenBlacklist[tokenID] = time.Now()
	tm.blacklistMutex.Unlock()
}

// RevokeRefreshToken removes refresh token from database
func (tm *TokenManager) RevokeRefreshToken(tokenID string) {
	tm.refreshDBMutex.Lock()
	delete(tm.refreshTokenDB, tokenID)
	tm.refreshDBMutex.Unlock()
}

// RefreshAccessToken generates new access token from refresh token
func (tm *TokenManager) RefreshAccessToken(refreshToken string) (string, string, error) {
	claims, err := tm.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new CSRF token for new access token
	csrfToken, err := tm.GenerateCSRFToken()
	if err != nil {
		return "", "", err
	}

	// Generate new access token
	newAccessToken, err := tm.GenerateAccessToken(claims.Username, csrfToken)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, csrfToken, nil
}

// cleanupExpiredTokens periodically removes expired tokens from blacklist
func (tm *TokenManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(tm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		tm.blacklistMutex.Lock()
		now := time.Now()
		for jti, revokedAt := range tm.tokenBlacklist {
			// Remove if older than the longest token TTL
			if now.Sub(revokedAt) > tm.config.AccessTokenTTL {
				delete(tm.tokenBlacklist, jti)
			}
		}
		tm.blacklistMutex.Unlock()

		// Also cleanup expired refresh tokens
		tm.refreshDBMutex.Lock()
		for jti, tokenData := range tm.refreshTokenDB {
			if now.After(tokenData.ExpiresAt) {
				delete(tm.refreshTokenDB, jti)
			}
		}
		tm.refreshDBMutex.Unlock()
	}
}

// Helper types
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	CSRFToken    string `json:"csrf_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Helper functions
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func generateJTI() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
