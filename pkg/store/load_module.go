package store

import (
	"context"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/pkg/loader"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func (s *SandboxStore) loadModule(moduleId string) (*ActiveModule, error) {
	s.mu.RLock()
	active, exists := s.activeModules[moduleId]
	if exists {
		active.wg.Add(1)
		s.mu.RUnlock()
		return active, nil
	}
	s.mu.RUnlock()

	s.loadingModulesMu.Lock()
	signal, exists := s.loadingModules[moduleId]

	if exists {
		s.loadingModulesMu.Unlock()
		<-signal
		return s.loadModule(moduleId)
	}

	s.loadingModules[moduleId] = make(chan struct{})
	s.loadingModulesMu.Unlock()

	defer func() {
		s.loadingModulesMu.Lock()
		defer s.loadingModulesMu.Unlock()
		if ch, ok := s.loadingModules[moduleId]; ok {
			close(ch)
			delete(s.loadingModules, moduleId)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Compile the module once
	compiled, err := loader.LoadCompiled(ctx, s.runtime, moduleId)
	if err != nil {
		return nil, err
	}

	// Instantiate a pool of instances from the compiled module
	instances := make(chan api.Module, s.poolSize)
	for range s.poolSize {
		inst, err := s.runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName(""))
		if err != nil {
			return nil, err
		}
		instances <- inst
	}

	mod := &ActiveModule{
		compiled:   compiled,
		instances:  instances,
		instanceId: moduleId,
	}
	mod.lastUsed.Store(time.Now().UnixNano())

	s.mu.Lock()
	if len(s.activeModules) >= int(s.maxActiveModules) {
		s.evictLRU()
	}
	s.activeModules[moduleId] = mod
	mod.wg.Add(1)
	s.mu.Unlock()

	return mod, nil
}
