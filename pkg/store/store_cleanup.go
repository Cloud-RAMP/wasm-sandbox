package store

import (
	"context"
	"log/slog"
	"time"
)

// Close all modules and remove them from the map
func (s *SandboxStore) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// remove all active modules from the map and close them
	for id, active := range s.activeModules {
		delete(s.activeModules, id)
		active.Close(s.poolSize)
	}

	// close the host module as well
	if s.hostModule != nil {
		s.hostModule.Close(ctx)
	}

	return s.runtime.Close(ctx)
}

// Shut down a singular module
//
// Remove all from the pool and close them, and finally close the compiled instance
func (mod *ActiveModule) Close(poolSize uint8) {
	mod.wg.Wait()

	for range poolSize {
		inst := <-mod.instances
		inst.Close(context.Background())
	}
	mod.compiled.Close(context.Background())
	close(mod.instances)
}

// Evict the least recently used module from the cache
func (s *SandboxStore) evictLRU() {
	var lru string
	var oldestTime *time.Time

	for mod := range s.activeModules {
		t := time.Unix(0, s.activeModules[mod].lastUsed.Load())
		if oldestTime == nil || t.Before(*oldestTime) {
			oldestTime = &t
			lru = mod
		}
	}

	// remove module from the map so no new requests can be sent to it
	mod := s.activeModules[lru]
	delete(s.activeModules, lru)

	go mod.Close(s.poolSize)
}

func (s *SandboxStore) cleanupIdleModules() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, active := range s.activeModules {
		t := time.Unix(0, active.lastUsed.Load())
		if time.Since(t) > s.maxIdleTime {
			slog.Info("Removing idle store", "storeId", id)
			delete(s.activeModules, id)
			go active.Close(s.poolSize)
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
