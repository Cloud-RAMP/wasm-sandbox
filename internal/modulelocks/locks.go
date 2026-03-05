// defines an interface that we can use to access the locks for a module's memory.
// this allows us to avoid errors with concurrent writes to memory.
//
// this done in an external package and not within the "store/ActiveModule" struct
// because it must be accessed in the host-builder/handler functions, which are passed
// a predetermined set of parameters that wazero chose.
//
// therefore, we need to define locks for our modules in a separate package, so that
// they can be used from multiple (loader, host-builder) while avoiding circular import issues.
//
// this package basically just manages references to locks that are to be used
// in other packages.
package modulelocks

import "sync"

type moduleLocks struct {
	locks map[string]*sync.Mutex
	mu    sync.RWMutex
}

var locks moduleLocks

func init() {
	locks.locks = make(map[string]*sync.Mutex)
}

func Delete(instanceId string) {
	locks.mu.Lock()
	delete(locks.locks, instanceId)
	locks.mu.Unlock()
}

func GetLockReference(instanceId string) *sync.Mutex {
	locks.mu.RLock()
	lock, ok := locks.locks[instanceId]
	if !ok {
		lock = &sync.Mutex{}
		locks.locks[instanceId] = lock
	}

	locks.mu.RUnlock()
	return lock
}
