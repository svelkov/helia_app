// global/vars.go
package global

import (
	"helia/config"
	"helia/internal/domain"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	gnFirma     string
	gnGod       int
	gnKar       int
	gnDuzSin    int
	gnLanguage  string
	mu          sync.RWMutex
	cfg         config.Config
	headerLocks sync.Map // Global map for managing resource locks by ID
)

func SetGnFirma(firma string) {
	mu.Lock()
	defer mu.Unlock()
	gnFirma = firma
}

func SetGnGod(god int) {
	mu.Lock()
	defer mu.Unlock()
	gnGod = god
}

func SetGnKar(kar int) {
	mu.Lock()
	defer mu.Unlock()
	gnKar = kar
}

func SetGnDuzSin(duzsin int) {
	mu.Lock()
	defer mu.Unlock()
	gnDuzSin = duzsin
}

func SetConfig(config config.Config) {
	mu.Lock()
	defer mu.Unlock()
	cfg = config
}
func SetGnLanguage(language string) {
	mu.Lock()
	defer mu.Unlock()
	gnLanguage = language
}

func GetGnFirma() string {
	mu.RLock()
	defer mu.RUnlock()
	return gnFirma
}

func GetGnGod() int {
	mu.RLock()
	defer mu.RUnlock()
	return gnGod
}

func GetGnKar() int {
	mu.RLock()
	defer mu.RUnlock()
	return gnKar
}

func GetGnDuzSin() int {
	mu.RLock()
	defer mu.RUnlock()
	return gnDuzSin
}

func GetConfig() config.Config {
	mu.RLock()
	defer mu.RUnlock()
	return cfg
}
func GetLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return gnLanguage
}

// GetHeaderLock retrieves or creates a mutex for the given resource ID
func GetHeaderLock(resourceID interface{}) *sync.Mutex {
	mu, _ := headerLocks.LoadOrStore(resourceID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// GetCurrentUser retrieves the username of the logged-in user from the request context
func GetCurrentUser(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}

// GetCurrentUserClaims retrieves the full UserClaims from the request context
func GetCurrentUserClaims(c *gin.Context) *domain.UserClaims {
	claims, exists := c.Get("userClaims")
	if !exists {
		return nil
	}

	userClaims, ok := claims.(*domain.UserClaims)
	if !ok {
		return nil
	}

	return userClaims
}
