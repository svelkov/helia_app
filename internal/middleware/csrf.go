package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const csrfTokenKey = "csrf_token"

// CSRFMiddleware protects against CSRF attacks
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
			formToken := c.PostForm("_csrf")

			session := sessions.Default(c)
			sessionToken := session.Get(csrfTokenKey)

			if formToken == "" || sessionToken == nil || formToken != sessionToken.(string) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
				c.Abort()
				return
			}

			// Generate new token for next request (single-use pattern)
			// but allow current handler to complete
			c.Next()

			// Clear token after handler completes to prevent reuse across requests
			session.Delete(csrfTokenKey)
			session.Save()
			return
		}

		c.Next()
	}
}
