package config

import (
	"log"
	"net/http"
)

type Router struct {
	Router *http.ServeMux
	Config Config
	Logger *log.Logger
	Auth   Authenticator
}

// Config struct for passing configuration
type Config struct {
	BasePath string
	God      int
	Kar      int
	PageSize int
}

func (c *Config) GetPageSize() int {
	if c.PageSize == 0 {
		return 20 // default page size
	}
	return c.PageSize
}

// Authenticator interface for authentication
type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type App struct {
	Cfg *Config
	Ws  *Router
}
