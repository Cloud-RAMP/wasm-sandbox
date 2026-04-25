package modulelocks

import "sync"

type entry struct {
	mu   sync.RWMutex
	refs int
}

type moduleLocks struct {
	entries map[string]*entry
	mu      sync.Mutex // protects entries map only
}

var lockTable = moduleLocks{
	entries: make(map[string]*entry),
}

func acquire(instanceId string) *entry {
	lockTable.mu.Lock()
	e, ok := lockTable.entries[instanceId]
	if !ok {
		e = &entry{}
		lockTable.entries[instanceId] = e
	}
	e.refs++
	lockTable.mu.Unlock()
	return e
}

func release(instanceId string) {
	lockTable.mu.Lock()
	e := lockTable.entries[instanceId]
	e.refs--
	if e.refs == 0 {
		delete(lockTable.entries, instanceId)
	}
	lockTable.mu.Unlock()
}

func Lock(instanceId string) {
	e := acquire(instanceId)
	e.mu.Lock()
}

func Unlock(instanceId string) {
	lockTable.mu.Lock()
	e := lockTable.entries[instanceId]
	lockTable.mu.Unlock()

	e.mu.Unlock()
	release(instanceId)
}

func RLock(instanceId string) {
	e := acquire(instanceId)
	e.mu.RLock()
}

func RUnlock(instanceId string) {
	lockTable.mu.Lock()
	e := lockTable.entries[instanceId]
	lockTable.mu.Unlock()

	e.mu.RUnlock()
	release(instanceId)
}

func Delete(instanceId string) {
	// Delete is now a no-op — release() cleans up automatically
	// when the last goroutine is done with the lock
}
