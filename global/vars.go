// global/vars.go
package global

import "sync"

var (
	gnGod int
	gnKar int
	mu    sync.RWMutex
)

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
