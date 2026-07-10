package domain

import (
	"context"

	"github.com/gin-gonic/gin"
)

// UserSession represents user-specific settings for the current request
// This replaces the global variables gnGod, gnKar, gnDuzSin, gnLanguage
// Each user maintains their own session preferences
type UserSession struct {
	UserID      int64  `json:"user_id"`      // From JWT claims
	UserName    string `json:"username"`     // From JWT claims
	Firma       string `json:"firma"`        // From JWT claims (company identifier)
	Mesto       string `json:"mesto"`        // User's city (for report headers)
	SelectedGod int    `json:"selected_god"` // Fiscal year (user-mutable - can change per session)
	SelectedKar int    `json:"selected_kar"` // Accounting period (user-mutable - can change per session)
	Language    string `json:"language"`     // UI language preference (user-mutable)
}

// GetSessionFromContext extracts UserSession from gin context
// Returns nil if not found (user not authenticated)
// Usage in service layer:
//
//	session := domain.GetSessionFromContext(c)
//	if session == nil {
//	    return errors.New("user session not found")
//	}
//	god := session.SelectedGod
func GetSessionFromContext(c *gin.Context) *UserSession {
	session, exists := c.Get("userSession")
	if !exists {
		return nil
	}

	userSession, ok := session.(*UserSession)
	if !ok {
		return nil
	}

	return userSession
}

// MustGetSessionFromContext extracts UserSession from context and panics if not found
// Use this when you're certain the session must exist (after auth middleware)
func MustGetSessionFromContext(c *gin.Context) *UserSession {
	session := GetSessionFromContext(c)
	if session == nil {
		panic("user session not found in context - auth middleware may not be configured")
	}
	return session
}

// SetSessionInContext stores UserSession in gin context
// Called by session middleware during request processing
func SetSessionInContext(c *gin.Context, session *UserSession) {
	c.Set("userSession", session)
}

// GetCurrentUserClaims retrieves the full UserClaims from the request context
func GetCurrentUserClaims(c *gin.Context) *UserClaims {
	claims, exists := c.Get("userClaims")
	if !exists {
		return nil
	}

	userClaims, ok := claims.(*UserClaims)
	if !ok {
		return nil
	}

	return userClaims
}

// ==================== context.Context Helpers (for framework-agnostic layers) ====================

type contextKey string

const UserSessionContextKey contextKey = "userSession"

// GetSessionFromStdContext retrieves UserSession from standard context.Context
// Use this in service/repository layers (framework-independent code)
// Returns nil if not found
func GetSessionFromStdContext(ctx context.Context) *UserSession {
	session, ok := ctx.Value(UserSessionContextKey).(*UserSession)
	if !ok {
		return nil
	}
	return session
}

// SetSessionInStdContext stores UserSession in standard context.Context
// Use this in handlers before passing context to service layer
// Example: ctx := domain.SetSessionInStdContext(c.Request.Context(), userSession)
func SetSessionInStdContext(ctx context.Context, session *UserSession) context.Context {
	return context.WithValue(ctx, UserSessionContextKey, session)
}
