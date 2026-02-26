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

// Removes a module from the sandbox store and closes it
func (s *SandboxStore) removeModule(ctx context.Context, moduleId string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()
	module, exists := s.activeModules[moduleId]
	if !exists {
		s.mu.Unlock()
		return nil
	} else {
		// remove module from active so that nobody can access it
		delete(s.activeModules, moduleId)
	}
	s.mu.Unlock()

	// TODO: Synchronization error.
	// Add a sync.WG to this runtime so that we can wait for all requests it is processing to finish
	return module.module.Close(ctx)
}

func (s *SandboxStore) cleanupIdleModules() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, active := range s.activeModules {
		if time.Since(active.lastUsed) > s.maxIdleTime {
			active.module.Close(context.Background())
			delete(s.activeModules, id)
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
