package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/loader"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type SandboxStore struct {
	runtime          wazero.Runtime
	hostModule       api.Module
	moduleConfig     wazero.ModuleConfig
	activeModules    map[string]*ActiveModule
	maxActiveModules uint16
	maxIdleTime      time.Duration
	cleanupInterval  time.Duration
	mu               sync.RWMutex // rw mutex might not be entirely necessary
	handlerMap       wasmevents.HandlerMap
}

type ActiveModule struct {
	module     api.Module
	lastUsed   time.Time
	instanceId string
}

type SandboxStoreCfg struct {
	MemoryLimitPages   uint32
	MaxActicveModules  uint16
	MaxExecutionTime   time.Duration
	CloseOnContextDone bool
	MaxIdleTime        time.Duration
	CleanupInterval    time.Duration
	HandlerMap         *wasmevents.HandlerMap
	Ctx                context.Context
}

// Execute a function on a given module
//
// The event will be handled by whatever custom event handler the user has set up
func (s *SandboxStore) ExecuteOnModule(ctx context.Context, wsEvent wsevents.WSEventInfo) error {
	if !wsEvent.EventType.Valid() {
		return fmt.Errorf("Invalid WS event type")
	}

	var active *ActiveModule
	s.mu.RLock()
	active, exists := s.activeModules[wsEvent.InstanceId]
	s.mu.RUnlock()

	// should be loaded from some external store
	if !exists {
		fmt.Println("Module does not exist, loading", wsEvent.InstanceId)

		// this var is done so that we don't use the := to make a local variable
		// so outer "active" can be assigned to return of LoadModule
		var err error
		active, err = s.loadModule(wsEvent.InstanceId)
		if err != nil {
			fmt.Println("Module loading failed!")
			return err
		}
	}

	if active == nil {
		fmt.Printf("Active is nil after loading")
		return fmt.Errorf("Active is nil after loading")
	}

	// create inner context with instanceId key / value
	ctx = context.WithValue(ctx, "instanceId", wsEvent.InstanceId)
	ctx = context.WithValue(ctx, "connectionId", wsEvent.ConnectionId)
	ctx = context.WithValue(ctx, "roomId", wsEvent.RoomId)

	// write the information of the event in module memory so they can read it
	ptr, memLen, err := asmscript.WriteWSEvent(active.module, wsEvent)

	// Call the `onMessage` function with the pointer and length
	onMessage := active.module.ExportedFunction(wsEvent.EventType.String())
	_, err = onMessage.Call(ctx, ptr, memLen)
	if err != nil {
		fmt.Printf("%s failed: %v\n", wsEvent.EventType.String(), err)
		return err
	}

	return nil
}

// Loads a given module into the sandbox store
//
// Returns an error if fetching the module failed
func (s *SandboxStore) loadModule(moduleId string) (*ActiveModule, error) {
	module, err := loader.Load(context.Background(), s.runtime, moduleId)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Add module to map of active modules
	s.activeModules[moduleId] = &ActiveModule{
		module:     module,
		lastUsed:   time.Now(),
		instanceId: moduleId,
	}

	return s.activeModules[moduleId], nil
}
