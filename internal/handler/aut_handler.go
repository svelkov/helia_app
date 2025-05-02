package handler

import (
	"encoding/json"
	"helia/internal/domain"
	"helia/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": accessToken, "refreshToken": refreshToken})
}

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request) {
	// The middleware should have already verified the token and added user claims to the context.
	claims, ok := r.Context().Value("userClaims").(*domain.UserClaims) // Assuming your claims are in domain.

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Access claims and return protected data.
	response := map[string]string{"message": "Protected resource", "username": claims.Username}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
