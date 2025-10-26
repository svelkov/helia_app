package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"helia/internal/i18n"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// Secret key for signing the JWT (keep this secure!)
	JwtSecret = []byte("3285f0d71eed0c41fded2115c9cc8ac09a0ab5a519565df10afdb20f8013c5268f2c19948b6af096c1cfc2921ab086be21fa5407b9d91aeb08eeeeef3c2e16c9ae30ae15f27d340f17c450468fef50795e58bb7351a94602bc045aea1a1ff3b03039081208cf067b44fd913b98b712e34ba080941f5ff8545b0eac26824f0ef4a93109939d8f917e1fac1eb588f4272ebac415975bcdc994c3a0fea7c3805d601443ad71dd9043858de5c2bfe64106683d9eaebce28442ce7bb22298d5b85cc3cc41e6f81f9c0f8f678cce559f745645edc5a5009ba20f8b5a16be4ee7dada7791913c90e3629a44b88a17d3d107bd3a6c0f3000b4865b2c015c0875901a028e")
)

// I18n middleware - detects and sets language
func I18n(translator *i18n.Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := getLangFromRequest(c)
		c.Set("lang", lang)
		c.Next()
	}
}

func getLangFromRequest(c *gin.Context) string {
	// 1. Check query parameter
	if lang := c.Query("lang"); lang != "" {
		// Set cookie for future requests
		c.SetCookie("lang", lang, 86400*30, "/", "", false, false)
		return lang
	}

	// 2. Check cookie
	if lang, err := c.Cookie("lang"); err == nil && lang != "" {
		return lang
	}

	// 3. Check Accept-Language header
	acceptLang := c.GetHeader("Accept-Language")
	if acceptLang != "" {
		langs := strings.Split(acceptLang, ",")
		if len(langs) > 0 {
			lang := strings.Split(langs[0], ";")[0]
			lang = strings.Split(lang, "-")[0]
			return lang
		}
	}

	// 4. Default language
	return "en"
}
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for these public routes
		publicPaths := []string{"/login", "/register", "/frontend/static"}
		for _, path := range publicPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}
		
		// Get token from cookie
		cookie, err := c.Cookie("auth_token")
		if err != nil {
			gin.DefaultWriter.Write([]byte(fmt.Sprintf("No auth token cookie found: %v", err)))
			c.Redirect(http.StatusSeeOther, "/login")
			//c.Abort()
			return
		}

		// Parse and validate token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(cookie, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return JwtSecret, nil
		})

		if err != nil {
			gin.DefaultWriter.Write([]byte(fmt.Sprintf("Token validation failed: %v", err)))
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		if !token.Valid {
			gin.DefaultWriter.Write([]byte("Invalid token"))
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Extract username and validate claims
		username, ok := claims["username"].(string)
		if !ok || username == "" {
			gin.DefaultWriter.Write([]byte("Username claim missing or invalid"))
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Check token expiration
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				gin.DefaultWriter.Write([]byte("Token expired"))
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}
		}

		// Add username to Gin context
		c.Set("username", username)
		c.Next()
	}
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger middleware - custom logging
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Simple colored logging
		statusColor := getStatusColor(statusCode)
		methodColor := getMethodColor(method)

		gin.DefaultWriter.Write([]byte(
			statusColor + " " +
				methodColor + " " +
				path + " " +
				clientIP + " " +
				latency.String() + "\n",
		))
	}
}

func getStatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "✓"
	case code >= 300 && code < 400:
		return "→"
	case code >= 400 && code < 500:
		return "⚠"
	default:
		return "✗"
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "[GET]"
	case "POST":
		return "[POST]"
	case "PUT":
		return "[PUT]"
	case "PATCH":
		return "[PATCH]"
	case "DELETE":
		return "[DELETE]"
	default:
		return "[" + method + "]"
	}
}
