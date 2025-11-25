// global/vars.go
package global

import (
	"helia/config"
	"sync"
)

var (
	gnFirma    string
	gnGod      int
	gnKar      int
	gnLanguage string
	mu         sync.RWMutex
	cfg        config.Config
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
