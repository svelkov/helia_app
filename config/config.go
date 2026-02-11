package config

import (
	"log"
	"net/http"
	"time"
)

type Router struct {
	Router *http.ServeMux
	Config Config
	Logger *log.Logger
	Auth   Authenticator
}

// Config struct for passing configuration
type Config struct {
	BasePath        string        `json:"base_path"`
	DBConfig        DB_Connection `json:"db_connection"`
	PageSize        int           `json:"page_size"`
	PageSizes       []int         `json:"page_sizes"`
	Env             string        `json:"env"`
	Port            string        `json:"port"`
	Languages       []string      `json:"languages"`
	DefaultLanguage string        `json:"defaultLanguage"`
	JwtSecret       string        `json:"jwt_secret"`
	SessionSecret   string        `json:"session_secret"`
	NDuzSint        int           `json:"nDuzSint"`

	// JWT Token Configuration (Option B - JWT-only authentication)
	AccessTokenTTL       int  `json:"access_token_ttl"`       // In minutes, default 15
	RefreshTokenTTL      int  `json:"refresh_token_ttl"`      // In minutes, default 1440 (24 hours)
	CSRFTokenTTL         int  `json:"csrf_token_ttl"`         // In minutes, default 1440
	EnableTokenBlacklist bool `json:"enable_token_blacklist"` // Enable token revocation, default true
}

type DB_Connection struct {
	DBHost       string `json:"db_host"`
	DBPort       int    `json:"db_port"`
	DBUser       string `json:"db_user"`
	DBPassword   string `json:"db_password"`
	DBName       string `json:"db_name"`
	DBSearchPath string `json:"db_search_path"`
}

func (c *Config) GetPageSize() int {
	if c.PageSize == 0 {
		return 20 // default page size
	}
	return c.PageSize
}

// GetAccessTokenTTL returns access token TTL with default fallback
func (c *Config) GetAccessTokenTTL() time.Duration {
	if c.AccessTokenTTL <= 0 {
		return 15 * time.Minute // default 15 minutes
	}
	return time.Duration(c.AccessTokenTTL) * time.Minute
}

// GetRefreshTokenTTL returns refresh token TTL with default fallback
func (c *Config) GetRefreshTokenTTL() time.Duration {
	if c.RefreshTokenTTL <= 0 {
		return 24 * time.Hour // default 24 hours
	}
	return time.Duration(c.RefreshTokenTTL) * time.Minute
}

// GetCSRFTokenTTL returns CSRF token TTL with default fallback
func (c *Config) GetCSRFTokenTTL() time.Duration {
	if c.CSRFTokenTTL <= 0 {
		return 24 * time.Hour // default 24 hours
	}
	return time.Duration(c.CSRFTokenTTL) * time.Minute
}

// IsTokenBlacklistEnabled returns whether token blacklist is enabled
func (c *Config) IsTokenBlacklistEnabled() bool {
	return c.EnableTokenBlacklist
}

// Authenticator interface for authentication
type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type App struct {
	Cfg *Config
	Ws  *Router
}
