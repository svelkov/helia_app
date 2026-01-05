package infrastructure

import (
	"testing"
	"time"
)

// TestTokenGeneration tests token pair generation
func TestTokenGeneration(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: true,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)
	username := "testuser"

	// Generate token pair
	pair, err := tm.GenerateTokenPair(username)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	// Verify all tokens are non-empty
	if pair.AccessToken == "" {
		t.Error("access token is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("refresh token is empty")
	}
	if pair.CSRFToken == "" {
		t.Error("CSRF token is empty")
	}

	// Verify expiry info
	if pair.ExpiresIn != 15*60 {
		t.Errorf("expected expires_in to be 900, got %d", pair.ExpiresIn)
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token_type to be Bearer, got %s", pair.TokenType)
	}
}

// TestAccessTokenVerification tests access token verification
func TestAccessTokenVerification(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: false,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)
	username := "testuser"

	pair, err := tm.GenerateTokenPair(username)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	// Verify access token
	claims, err := tm.VerifyAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("failed to verify access token: %v", err)
	}

	if claims.Username != username {
		t.Errorf("expected username %s, got %s", username, claims.Username)
	}

	// Token should have a subject
	if claims.Subject != username {
		t.Errorf("expected subject %s, got %s", username, claims.Subject)
	}

	// Token should have an ID
	if claims.ID == "" {
		t.Error("token ID should not be empty")
	}
}

// TestCSRFTokenVerification tests CSRF token verification
func TestCSRFTokenVerification(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: false,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)
	username := "testuser"

	pair, err := tm.GenerateTokenPair(username)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	// Verify CSRF token matches
	valid, err := tm.VerifyCSRFToken(pair.CSRFToken, pair.AccessToken)
	if err != nil {
		t.Fatalf("CSRF verification error: %v", err)
	}

	if !valid {
		t.Error("CSRF token verification should return true")
	}

	// Verify mismatched CSRF token fails
	validWrong, err := tm.VerifyCSRFToken("wrong-csrf-token", pair.AccessToken)
	if err != nil {
		t.Fatalf("CSRF verification error: %v", err)
	}

	if validWrong {
		t.Error("mismatched CSRF token should return false")
	}
}

// TestRefreshTokenVerification tests refresh token verification
func TestRefreshTokenVerification(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: false,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)
	username := "testuser"

	pair, err := tm.GenerateTokenPair(username)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	// Verify refresh token
	claims, err := tm.VerifyRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("failed to verify refresh token: %v", err)
	}

	if claims.Username != username {
		t.Errorf("expected username %s, got %s", username, claims.Username)
	}
}

// TestTokenRevocation tests token blacklisting
func TestTokenRevocation(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: true,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)
	username := "testuser"

	pair, err := tm.GenerateTokenPair(username)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	// Verify token works initially
	claims, err := tm.VerifyAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("failed to verify fresh token: %v", err)
	}

	// Revoke the token
	tm.RevokeAccessToken(claims.ID)

	// Verify token is now blacklisted
	_, err = tm.VerifyAccessToken(pair.AccessToken)
	if err == nil {
		t.Error("revoked token should fail verification")
	}

	// Verify error contains "revoked" message
	if err.Error() != "token has been revoked" {
		t.Errorf("expected 'token has been revoked' error, got: %v", err)
	}
}

// TestAccessTokenRefresh tests refreshing an access token
func TestAccessTokenRefresh(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: false,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)
	username := "testuser"

	pair, err := tm.GenerateTokenPair(username)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	// Store original tokens
	originalAccessToken := pair.AccessToken
	originalCSRFToken := pair.CSRFToken

	// Refresh access token
	newAccessToken, newCSRFToken, err := tm.RefreshAccessToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("failed to refresh access token: %v", err)
	}

	// Verify new tokens are different
	if newAccessToken == originalAccessToken {
		t.Error("new access token should be different from original")
	}

	if newCSRFToken == originalCSRFToken {
		t.Error("new CSRF token should be different from original")
	}

	// Verify new access token is valid
	claims, err := tm.VerifyAccessToken(newAccessToken)
	if err != nil {
		t.Fatalf("failed to verify refreshed token: %v", err)
	}

	if claims.Username != username {
		t.Errorf("expected username %s, got %s", username, claims.Username)
	}

	// Verify CSRF token works with new access token
	valid, err := tm.VerifyCSRFToken(newCSRFToken, newAccessToken)
	if err != nil {
		t.Fatalf("CSRF verification error: %v", err)
	}

	if !valid {
		t.Error("new CSRF token should verify successfully")
	}
}

// TestTokenWithDifferentSecrets tests that different secrets invalidate tokens
func TestTokenWithDifferentSecrets(t *testing.T) {
	secret1 := []byte("test-secret-key-min-32-chars!!!")
	secret2 := []byte("different-secret-key-min-32!!!!")

	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: false,
		CleanupInterval: 1 * time.Hour,
	}

	tm1 := NewTokenManager(secret1, config)
	tm2 := NewTokenManager(secret2, config)

	pair, err := tm1.GenerateTokenPair("testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Token should verify with correct secret
	_, err = tm1.VerifyAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("token should verify with correct secret: %v", err)
	}

	// Token should NOT verify with wrong secret
	_, err = tm2.VerifyAccessToken(pair.AccessToken)
	if err == nil {
		t.Error("token should not verify with wrong secret")
	}
}

// TestTokenExpiration tests token expiration behavior
func TestTokenExpiration(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  1 * time.Millisecond, // Very short TTL for testing
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: false,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)

	pair, err := tm.GenerateTokenPair("testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Token should be expired
	_, err = tm.VerifyAccessToken(pair.AccessToken)
	if err == nil {
		t.Error("expired token should fail verification")
	}
}

// TestGenerateCSRFToken tests standalone CSRF token generation
func TestGenerateCSRFToken(t *testing.T) {
	secret := []byte("test-secret-key-min-32-chars!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: false,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)

	// Generate multiple CSRF tokens
	tokens := make(map[string]bool)
	for i := 0; i < 10; i++ {
		token, err := tm.GenerateCSRFToken()
		if err != nil {
			t.Fatalf("failed to generate CSRF token: %v", err)
		}

		if token == "" {
			t.Error("CSRF token should not be empty")
		}

		if tokens[token] {
			t.Error("CSRF tokens should be unique")
		}
		tokens[token] = true
	}

	if len(tokens) != 10 {
		t.Errorf("expected 10 unique tokens, got %d", len(tokens))
	}
}

// Example usage demonstrating complete flow
func ExampleTokenManager() {
	secret := []byte("your-secret-key-min-32-chars-long!!!")
	config := TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		CSRFTokenTTL:    24 * time.Hour,
		EnableBlacklist: true,
		CleanupInterval: 1 * time.Hour,
	}

	tm := NewTokenManager(secret, config)

	// 1. Generate tokens for login
	tokenPair, _ := tm.GenerateTokenPair("user@example.com")
	println("Access Token:", tokenPair.AccessToken)
	println("Refresh Token:", tokenPair.RefreshToken)
	println("CSRF Token:", tokenPair.CSRFToken)

	// 2. Verify access token and claims
	claims, _ := tm.VerifyAccessToken(tokenPair.AccessToken)
	println("Username:", claims.Username)

	// 3. Verify CSRF token
	valid, _ := tm.VerifyCSRFToken(tokenPair.CSRFToken, tokenPair.AccessToken)
	println("CSRF Valid:", valid)

	// 4. Refresh access token on expiration
	newAccess, newCSRF, _ := tm.RefreshAccessToken(tokenPair.RefreshToken)
	println("New Access Token:", newAccess)
	println("New CSRF Token:", newCSRF)

	// 5. Revoke token on logout
	tm.RevokeAccessToken(claims.ID)
	println("Token revoked")

	// 6. Verify revoked token fails
	_, err := tm.VerifyAccessToken(tokenPair.AccessToken)
	println("Verification after revocation:", err.Error())
}
