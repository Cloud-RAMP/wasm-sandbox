package modulelocks

import (
	"sync"
)

type moduleLocks struct {
	locks map[string]*sync.Mutex
	mu    sync.RWMutex
}

var lockTable = moduleLocks{
	locks: make(map[string]*sync.Mutex),
}

func Lock(instanceId string) {
	lockTable.mu.RLock()
	lock, ok := lockTable.locks[instanceId]
	lockTable.mu.RUnlock()

	if ok {
		lock.Lock()
		return
	}

	lockTable.mu.Lock()
	defer lockTable.mu.Unlock()

	lock = &sync.Mutex{}
	lockTable.locks[instanceId] = lock
	lock.Lock()
}

func Unlock(instanceId string) {
	lockTable.mu.RLock()
	lock, ok := lockTable.locks[instanceId]
	lockTable.mu.RUnlock()

	if ok {
		lock.Unlock()
		return
	}

	lockTable.mu.Lock()
	defer lockTable.mu.Unlock()

	lock = &sync.Mutex{}
	lockTable.locks[instanceId] = lock
	lock.Unlock()
}

func Delete(instanceId string) *sync.Mutex {
	lockTable.mu.Lock()
	defer lockTable.mu.Unlock()

	if lock, ok := lockTable.locks[instanceId]; ok {
		delete(lockTable.locks, instanceId)
		return lock
	}
	return nil
}
