package handler

import (
	"encoding/json"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService  service.AuthService
	tokenManager *infrastructure.TokenManager
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// NewAuthHandlerWithTokenManager creates auth handler with token manager for Option B
func NewAuthHandlerWithTokenManager(authService service.AuthService, tokenManager *infrastructure.TokenManager) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		tokenManager: tokenManager,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	CSRFToken    string `json:"csrf_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
	CSRFToken   string `json:"csrf_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// Login handles user login with JWT token pair generation (Option B)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Generate token pair using TokenManager if available
	if h.tokenManager != nil {
		// Verify credentials with auth service (using Login which returns tokens)
		_, _, err := h.authService.Login(req.Username, req.Password)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenPair, err := h.tokenManager.GenerateTokenPair(req.Username)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(loginResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			CSRFToken:    tokenPair.CSRFToken,
			ExpiresIn:    tokenPair.ExpiresIn,
			TokenType:    tokenPair.TokenType,
		})
		return
	}

	// Fallback to legacy behavior if TokenManager not initialized
	accessToken, refreshToken, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// Refresh handles token refresh (Option B)
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if h.tokenManager == nil {
		http.Error(w, "Token refresh not enabled", http.StatusServiceUnavailable)
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	accessToken, csrfToken, err := h.tokenManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(refreshResponse{
		AccessToken: accessToken,
		CSRFToken:   csrfToken,
		ExpiresIn:   15 * 60, // 15 minutes in seconds
		TokenType:   "Bearer",
	})
}

// Logout handles user logout with token revocation (Option B)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.tokenManager == nil {
		http.Error(w, "Logout not available", http.StatusServiceUnavailable)
		return
	}

	claims, ok := r.Context().Value("userClaims").(*domain.UserClaims)
	if !ok || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Revoke the current access token
	if claims.ID != "" {
		h.tokenManager.RevokeAccessToken(claims.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// Protected is a helper to demonstrate protected endpoint access
func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("userClaims").(*domain.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]string{"message": "Protected resource", "username": claims.Username}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
