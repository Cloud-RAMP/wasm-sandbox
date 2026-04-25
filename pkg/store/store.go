package store

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/loader"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type SandboxStore struct {
	// The runtime is where modules are executed, but thier states are kept separete
	runtime wazero.Runtime

	// The host module is where we define go functions to be accessed within the WASM module.
	// See the hostbuilder module for more
	hostModule   api.Module
	moduleConfig wazero.ModuleConfig

	// A map between moduleId and the actual WASM modules
	// This could be refactored into its own type
	activeModules    map[string]*ActiveModule
	maxActiveModules uint16
	maxIdleTime      time.Duration
	cleanupInterval  time.Duration
	maxExecutionTime time.Duration
	poolSize         uint8
	mu               sync.RWMutex // rw mutex might not be entirely necessary

	// Map that the user of this package will need to instantiate.
	// It allows us to designate functions to run on every event, and will likely be used
	// for things like external data store fetching.
	handlerMap wasmevents.HandlerMap

	// loadingModules defines a set of chans which will be present when a module is actively being loaded.
	// This is done to prevent concurrent fetches of the same module
	loadingModules   map[string]chan struct{}
	loadingModulesMu sync.Mutex
}

type ActiveModule struct {
	// The compiled bytes of a WASM module
	compiled wazero.CompiledModule

	// A pool of running instances.
	//
	// Goroutines pull these off one at a time to avoid concurrent writes to memory
	instances chan api.Module

	// Atomic representation of when the module was last used
	lastUsed atomic.Int64

	// The module's ID
	instanceId string

	// Waitgroup so that we don't
	wg sync.WaitGroup
}

type SandboxStoreCfg struct {
	MemoryLimitPages   uint32
	MaxActiveModules   uint16
	MaxExecutionTime   time.Duration
	CloseOnContextDone bool
	MaxIdleTime        time.Duration
	CleanupInterval    time.Duration
	HandlerMap         *wasmevents.HandlerMap
	Ctx                context.Context
	LoaderFunction     loader.LoaderFunction
	PoolSize           uint8
}

// Execute a function on a given module
//
// The event will be handled by whatever custom event handler the user has set up
func (s *SandboxStore) ExecuteOnModule(ctx context.Context, wsEvent *wsevents.WSEventInfo) error {
	if !wsEvent.EventType.Valid() {
		return fmt.Errorf("Invalid WS event type")
	}

	active, err := s.loadModule(wsEvent.InstanceId)
	if err != nil {
		return err
	}
	if active == nil {
		return fmt.Errorf("Active is nil after loading")
	}

	// Increment waitgroup. This way we can wait on all modules to be returned before closing
	active.wg.Add(1)
	defer active.wg.Done()

	// Grab an instance from the pool — blocks if all instances are in use
	instance := <-active.instances
	defer func() { active.instances <- instance }()

	// Update last used
	active.lastUsed.Store(time.Now().UnixNano())

	// Create inner context with instanceId key / value
	ctx = context.WithValue(ctx, "instanceId", wsEvent.InstanceId)
	ctx = context.WithValue(ctx, "connectionId", wsEvent.ConnectionId)
	ctx = context.WithValue(ctx, "roomId", wsEvent.RoomId)

	// Add timeout (defaults to 5 seconds)
	ctx, cancel := context.WithTimeout(ctx, s.maxExecutionTime)
	defer cancel()

	// Write the information of the event in module memory so they can read it
	ptr, memLen, err := asmscript.WriteWSEvent(&asmscript.ModuleContext{
		Module: instance,
		Ctx:    ctx,
	}, wsEvent)
	if err != nil {
		return err
	}

	onMessage := instance.ExportedFunction(wsEvent.EventType.String())
	_, err = onMessage.Call(ctx, ptr, memLen)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
