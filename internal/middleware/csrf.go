package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const csrfTokenKey = "csrf_token"

// getCsrfTokenFromSessionOrContext retrieves CSRF token from session, fallback to context
func getCsrfTokenFromSessionOrContext(c *gin.Context) (string, bool) {
	session := sessions.Default(c)
	sessionToken := session.Get(csrfTokenKey)

	if sessionToken == nil {
		return "", false
	}

	if tokenStr, ok := sessionToken.(string); ok {
		return tokenStr, true
	}

	return "", false
}

// CSRFMiddleware protects against CSRF attacks
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF validation for public routes (login, register, static)
		publicPaths := []string{"/login", "/register", "/frontend/static"}
		for _, path := range publicPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// For GET/HEAD/OPTIONS requests, generate or return CSRF token
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			session := sessions.Default(c)
			token := session.Get(csrfTokenKey)

			// Generate new token if not exists
			if token == nil {
				token = uuid.New().String()
				session.Set(csrfTokenKey, token)
				session.Save()
			}
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// For POST/PUT/DELETE, validate CSRF token
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodDelete {
			// Try to get token from form data first
			formToken := c.PostForm("_csrf")

			// Get session token
			sessionToken, exists := getCsrfTokenFromSessionOrContext(c)

			// Debug logging
			if formToken == "" {
				fmt.Printf("CSRF validation failed: no form token provided for %s\n", c.Request.RequestURI)
			}
			if !exists {
				fmt.Printf("CSRF validation failed: no session token for %s\n", c.Request.RequestURI)
			}
			if exists && formToken != sessionToken {
				fmt.Printf("CSRF token mismatch for %s: form=%s, session=%s\n", c.Request.RequestURI, formToken, sessionToken)
			}

			if formToken == "" || !exists || formToken != sessionToken {
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
				c.Abort()
				return
			}

			// Token is valid - keep same token for session duration
			c.Set("csrf_token", sessionToken)

			c.Next()
			return
		}

		c.Next()
	}
}
