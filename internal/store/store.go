package store

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/builder"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/events"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/handlers"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type SandboxStore struct {
	runtime         wazero.Runtime
	hostModule      api.Module
	moduleConfig    wazero.ModuleConfig
	activeModules   map[string]*ActiveModule
	maxIdleTime     time.Duration
	cleanupInterval time.Duration
	mu              sync.RWMutex // rw mutex might not be entirely necessary
	handlerMap      events.HandlerMap
}

type ActiveModule struct {
	module     api.Module
	lastUsed   time.Time
	wasmBytes  []byte
	instanceId string
}

type SandboxStoreCfg struct {
	MemoryLimitPages   uint32
	CloseOnContextDone bool
	MaxIdleTime        time.Duration
	CleanupInterval    time.Duration
	HandlerMap         *events.HandlerMap
	Ctx                context.Context
}

// Will probably need to pass a ctx into this later, or limit execution time somehow
func NewSandboxStore(cfg SandboxStoreCfg) (*SandboxStore, error) {
	ctx := context.Background()

	memPages := cfg.MemoryLimitPages
	if memPages == 0 {
		memPages = 10
	}

	// Create runtime with limits
	runtime := wazero.NewRuntimeWithConfig(ctx,
		wazero.NewRuntimeConfig().
			WithMemoryLimitPages(memPages).
			WithCloseOnContextDone(cfg.CloseOnContextDone))

	// Build host module once
	hostModule, err := builder.BuildHostModule(runtime, cfg.HandlerMap)
	if err != nil {
		runtime.Close(ctx)
		return nil, err
	}

	store := &SandboxStore{
		runtime:       runtime,
		hostModule:    hostModule,
		moduleConfig:  wazero.NewModuleConfig(),
		activeModules: make(map[string]*ActiveModule),
	}

	// auto-clean up modules if cleanup interval and max idle time are defined
	if cfg.CleanupInterval != 0 && cfg.MaxIdleTime != 0 {
		store.cleanupInterval = cfg.CleanupInterval
		store.maxIdleTime = cfg.MaxIdleTime
		store.StartCleanupRoutine()
	}

	return store, nil
}

// Loads a given module into the sandbox store
//
// Returns an error if instantiation failed
func (s *SandboxStore) LoadModule(moduleId string, wasmBytes []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already loaded
	if _, exists := s.activeModules[moduleId]; exists {
		return nil
	}

	// Instantiate the module
	ctx := context.Background()
	module, err := s.runtime.Instantiate(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to instantiate module: %w", err)
	}

	// Add module to map of active modules
	s.activeModules[moduleId] = &ActiveModule{
		module:     module,
		lastUsed:   time.Now(),
		wasmBytes:  wasmBytes,
		instanceId: moduleId,
	}

	return nil
}

// Execute a function on a given module
func (s *SandboxStore) ExecuteOnModule(ctx context.Context, moduleId string, handler handlers.Handler, payload string) ([]events.Event, error) {
	s.mu.RLock()
	active, exists := s.activeModules[moduleId]
	s.mu.RUnlock()

	// should be loaded from some external store
	if !exists {
		return nil, fmt.Errorf("module %s not loaded", moduleId)
	}

	// create inner context with instanceId key / value
	ctx = context.WithValue(ctx, "instanceId", moduleId)

	// operate on the requested handler
	switch handler {
	case handlers.ON_MESSAGE:
		ptr, memLen, err := asmscript.CreateASString(active.module, payload)
		if err != nil {
			log.Fatalf("Failed to create WASM string %v", err)
			return nil, err
		}

		// Call the `onMessage` function with the pointer and length
		onMessage := active.module.ExportedFunction(handler.String())
		_, err = onMessage.Call(ctx, ptr, memLen)
		if err != nil {
			log.Fatalf("%s failed: %v", handler.String(), err)
		}
	default:
		fmt.Printf("Unimplemented handler! %v", handler)
		return nil, fmt.Errorf("Bad handler")
	}

	return nil, nil
}

// Removes a module from the sandbox store and closes it
func (s *SandboxStore) RemoveModule(ctx context.Context, moduleId string) error {
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

	return module.module.Close(ctx)
}

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

func (s *SandboxStore) CleanupIdleModules() {
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

func (s *SandboxStore) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(s.cleanupInterval)
		for range ticker.C {
			s.CleanupIdleModules()
		}
	}()
}
