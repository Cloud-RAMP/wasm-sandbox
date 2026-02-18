package modulestore

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
	runtime        wazero.Runtime
	hostModule     api.Module
	moduleConfig   wazero.ModuleConfig
	activeModules  map[string]*ActiveModule
	modulesMutex   sync.RWMutex // rw mutex might not be entirely necessary
	invocationData InvocationData
	dataMutex      sync.Mutex
}

type ActiveModule struct {
	module     api.Module
	lastUsed   time.Time
	wasmBytes  []byte
	instanceId string
}

type InvocationData struct {
	Events []events.Event
}

type SandboxStoreCfg struct {
	MemoryLimitPages   uint8
	CloseOnContextDone bool
	Ctx                context.Context
}

// Will probably need to pass a ctx into this later, or limit execution time somehow
func NewSandboxStore() (*SandboxStore, error) {
	ctx := context.Background()

	// Create runtime with limits
	runtime := wazero.NewRuntimeWithConfig(ctx,
		wazero.NewRuntimeConfig().
			WithMemoryLimitPages(10).
			WithCloseOnContextDone(true))

	// Build host module once
	hostModule, err := builder.BuildHostModule(runtime)
	if err != nil {
		runtime.Close(ctx)
		return nil, err
	}

	return &SandboxStore{
		runtime:       runtime,
		hostModule:    hostModule,
		moduleConfig:  wazero.NewModuleConfig(),
		activeModules: make(map[string]*ActiveModule),
	}, nil
}

// Loads a given module into the sandbox store
//
// Returns an error if instantiation failed
func (s *SandboxStore) LoadModule(moduleId string, wasmBytes []byte) error {
	s.modulesMutex.Lock()
	defer s.modulesMutex.Unlock()

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
	s.modulesMutex.RLock()
	active, exists := s.activeModules[moduleId]
	s.modulesMutex.RUnlock()

	// should be loaded from some external store
	if !exists {
		return nil, fmt.Errorf("module %s not loaded", moduleId)
	}

	// Reset invocation data for this execution
	s.dataMutex.Lock()
	s.invocationData = InvocationData{
		Events: []events.Event{},
	}
	s.dataMutex.Unlock()

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
		_, err = onMessage.Call(context.Background(), ptr, memLen)
		if err != nil {
			log.Fatalf("%s failed: %v", handler.String(), err)
		}
	default:
		fmt.Printf("Unimplemented handler! %v", handler)
		return nil, fmt.Errorf("Bad handler")
	}

	// Return captured events
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()
	return s.invocationData.Events, nil
}
