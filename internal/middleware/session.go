package middleware

import (
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// ContextWithSessionMiddleware enriches context.Context with UserSession from gin.Context
// This middleware should run AFTER auth middleware
// Now all handlers automatically get c.Request.Context() with UserSession already set
//
// Usage in router setup:
//
//	router.Use(middleware.Auth())  // Auth middleware first
//	router.Use(middleware.ContextWithSessionMiddleware())  // Then enrich context
//	router.GET("/api/endpoint", handler.Handler)  // Now context has userSession
func ContextWithSessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get userSession from gin.Context (set by auth middleware)
		userSession := domain.GetSessionFromContext(c)
		if userSession != nil {
			// Store in context.Context for service layer to use
			savedCtx := domain.SetSessionInStdContext(c.Request.Context(), userSession)
			// Replace the request's context with the enriched one
			c.Request = c.Request.WithContext(savedCtx)
		}
		c.Next()
	}
}
