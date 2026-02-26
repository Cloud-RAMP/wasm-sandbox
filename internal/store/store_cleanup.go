package store

import (
	"context"
	"log"
	"time"
)

// Close all modules and remove them from the map
func (s *SandboxStore) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, active := range s.activeModules {
		active.module.Close(ctx)
		delete(s.activeModules, id)
	}

	// close the host module as well
	if s.hostModule != nil {
		s.hostModule.Close(ctx)
	}

	return s.runtime.Close(ctx)
}

// Evict the least recently used module from the cache
//
// Assumes that the lock is held before the function runs
func (s *SandboxStore) evictLRU() {
	var lru string
	var oldestTime *time.Time
	for mod := range s.activeModules {
		if oldestTime == nil || s.activeModules[mod].lastUsed.Before(*oldestTime) {
			oldestTime = &s.activeModules[mod].lastUsed
			lru = mod
		}
	}

	// remove module from the map so no new requests can be sent to it
	mod := s.activeModules[lru]
	delete(s.activeModules, lru)

	// detatch a goroutine to wait on the module's requests and then close it
	go func() {
		mod.wg.Wait()
		mod.module.Close(context.Background())
	}()
}

func (s *SandboxStore) cleanupIdleModules() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, active := range s.activeModules {
		if time.Since(active.lastUsed) > s.maxIdleTime {
			delete(s.activeModules, id)
			active.wg.Wait()
			active.module.Close(context.Background())
			log.Printf("Unloaded idle module: %s", id)
		}
	}
}

func (s *SandboxStore) startCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(s.cleanupInterval)
		for range ticker.C {
			s.cleanupIdleModules()
		}
	}()
}
