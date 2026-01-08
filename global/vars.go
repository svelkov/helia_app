// global/vars.go
package global

import (
	"sync"
)

var (
	headerLocks sync.Map // Global map for managing resource locks by ID
)

// GetHeaderLock retrieves or creates a mutex for the given resource ID
func GetHeaderLock(resourceID interface{}) *sync.Mutex {
	lock, _ := headerLocks.LoadOrStore(resourceID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
