package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure"
)

// JWTAuthMiddleware provides Option B JWT-only authentication with integrated CSRF protection
type JWTAuthMiddleware struct {
	tokenManager  *infrastructure.TokenManager
	excludedPaths map[string]bool
}

// NewJWTAuthMiddleware creates a new JWT auth middleware
func NewJWTAuthMiddleware(tokenManager *infrastructure.TokenManager) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		tokenManager: tokenManager,
		excludedPaths: map[string]bool{
			"/login":            true,
			"/register":         true,
			"/health":           true,
			"/api/auth/login":   true,
			"/api/auth/refresh": true,
			"/frontend/static":  true,
		},
	}
}

// Handler wraps http.Handler for net/http mux
func (m *JWTAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is excluded from authentication
		if m.isPathExcluded(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Verify JWT token
		claims, err := m.verifyToken(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		// For non-safe methods (POST, PUT, DELETE, PATCH), verify CSRF token
		if !m.isSafeMethod(r.Method) {
			if err := m.verifyCSRF(r, claims); err != nil {
				http.Error(w, fmt.Sprintf("CSRF verification failed: %v", err), http.StatusForbidden)
				return
			}
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "userClaims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// verifyToken extracts and validates the JWT token from the request
func (m *JWTAuthMiddleware) verifyToken(r *http.Request) (*domain.UserClaims, error) {
	// Try to get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Parse "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	claims, err := m.tokenManager.VerifyAccessToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	return claims, nil
}

// verifyCSRF validates CSRF token against the JWT claims
func (m *JWTAuthMiddleware) verifyCSRF(r *http.Request, claims *domain.UserClaims) error {
	// Get CSRF token from request
	csrfToken := m.getCSRFToken(r)
	if csrfToken == "" {
		return fmt.Errorf("missing CSRF token")
	}

	// Get the original access token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid authorization header")
	}

	accessToken := parts[1]

	// Verify CSRF token against the hash in JWT claims
	valid, err := m.tokenManager.VerifyCSRFToken(csrfToken, accessToken)
	if err != nil {
		return fmt.Errorf("CSRF token verification error: %w", err)
	}

	if !valid {
		return fmt.Errorf("CSRF token does not match")
	}

	return nil
}

// getCSRFToken extracts CSRF token from multiple possible locations
func (m *JWTAuthMiddleware) getCSRFToken(r *http.Request) string {
	// 1. Check custom header (recommended for AJAX)
	if token := r.Header.Get("X-CSRF-Token"); token != "" {
		return token
	}

	// 2. Check Authorization-CSRF header (alternative)
	if token := r.Header.Get("X-CSRF-Authorization"); token != "" {
		return token
	}

	// 3. Check form data (for traditional form submissions)
	if err := r.ParseForm(); err == nil {
		if token := r.FormValue("csrf_token"); token != "" {
			return token
		}
	}

	// 4. Check query parameter (not recommended but supported)
	if token := r.URL.Query().Get("csrf_token"); token != "" {
		return token
	}

	return ""
}

// isSafeMethod checks if the HTTP method is safe (GET, HEAD, OPTIONS)
func (m *JWTAuthMiddleware) isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

// isPathExcluded checks if the path should be excluded from auth
func (m *JWTAuthMiddleware) isPathExcluded(path string) bool {
	// Direct match
	if m.excludedPaths[path] {
		return true
	}

	// Prefix match
	for excluded := range m.excludedPaths {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}

	return false
}

// AddExcludedPath adds a path to the exclusion list
func (m *JWTAuthMiddleware) AddExcludedPath(path string) {
	m.excludedPaths[path] = true
}

// CSRFTokenProvider provides CSRF token to templates and clients
// This middleware adds the CSRF token to response headers for easy access
type CSRFTokenProvider struct {
	tokenManager *infrastructure.TokenManager
}

// NewCSRFTokenProvider creates a new CSRF token provider
func NewCSRFTokenProvider(tokenManager *infrastructure.TokenManager) *CSRFTokenProvider {
	return &CSRFTokenProvider{tokenManager: tokenManager}
}

// Handler returns a middleware that provides CSRF tokens
func (p *CSRFTokenProvider) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get claims from context if they exist (from JWT auth middleware)
		claims, ok := r.Context().Value("userClaims").(*domain.UserClaims)
		if ok && claims != nil {
			// Generate a new CSRF token for the response
			csrfToken, err := p.tokenManager.GenerateCSRFToken()
			if err == nil {
				// Add to response header so frontend can read it
				w.Header().Set("X-CSRF-Token", csrfToken)
				// Also add to context for template rendering
				ctx := context.WithValue(r.Context(), "csrfToken", csrfToken)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
