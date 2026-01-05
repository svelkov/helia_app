package infrastructure

import "github.com/gin-gonic/gin"

type contextKey string

const usernameKey contextKey = "username"

// AuthMiddleware checks for a valid JWT and passes user data to context.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("auth_token")
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Get JWT secret from context
		var secret []byte
		if s, exists := c.Get("jwtSecret"); exists {
			if secretBytes, ok := s.([]byte); ok {
				secret = secretBytes
			}
		}

		if secret == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		username, err := VerifyJWT(tokenString, secret)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		c.Set("username", username)
		c.Next()
	}
}
