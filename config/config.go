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
	BasePath        string        `json:"base_path"`
	DBConfig        DB_Connection `json:"db_connection"`
	PageSize        int           `json:"page_size"`
	PageSizes       []int         `json:"page_sizes"`
	Env             string        `json:"env"`
	Port            string        `json:"port"`
	Languages       []string      `json:"languages"`
	DefaultLanguage string        `json:"defaultLanguage"`
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

// Authenticator interface for authentication
type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type App struct {
	Cfg *Config
	Ws  *Router
}
