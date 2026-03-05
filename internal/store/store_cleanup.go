package store

import (
	"context"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	modulelocks "github.com/Cloud-RAMP/wasm-sandbox/internal/module-locks"
)

// Close all modules and remove them from the map
func (s *SandboxStore) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	logging.Logger.Info("Closing sandbox store")

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

	logging.Logger.Infof("Removing LRU module %s", lru)

	// detatch a goroutine to wait on the module's requests and then close it
	go func() {
		mod.wg.Wait()
		modulelocks.Delete(lru)
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
			modulelocks.Delete(id)
			active.module.Close(context.Background())
			logging.Logger.Infof("Removing inactive module %s", id)
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
