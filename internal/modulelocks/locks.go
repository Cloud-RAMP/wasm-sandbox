package modulelocks

import "sync"

type moduleLocks struct {
	locks map[string]*sync.Mutex
	mu    sync.Mutex
}

var locks moduleLocks

func init() {
	locks.locks = make(map[string]*sync.Mutex)
}

func Lock(instanceId string) {
	locks.mu.Lock()
	lock, ok := locks.locks[instanceId]
	if !ok {
		lock = &sync.Mutex{}
		locks.locks[instanceId] = lock
	}
	lock.Lock()
	locks.mu.Unlock()
}

func Unlock(instanceId string) {
	locks.mu.Lock()
	lock, ok := locks.locks[instanceId]
	if ok {
		lock.Unlock()
	}
	locks.mu.Unlock()
}

func Delete(instanceId string) {
	locks.mu.Lock()
	delete(locks.locks, instanceId)
	locks.mu.Unlock()
}
