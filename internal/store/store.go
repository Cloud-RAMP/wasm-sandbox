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
	maxExecutionTime time.Duration
	mu               sync.RWMutex // rw mutex might not be entirely necessary
	handlerMap       wasmevents.HandlerMap

	loadingModulesMu sync.Mutex
	loadingModules   map[string]chan struct{}
}

type ActiveModule struct {
	module     api.Module
	lastUsed   time.Time
	instanceId string
	wg         sync.WaitGroup
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
}

// Execute a function on a given module
//
// The event will be handled by whatever custom event handler the user has set up
func (s *SandboxStore) ExecuteOnModule(ctx context.Context, wsEvent wsevents.WSEventInfo) error {
	if !wsEvent.EventType.Valid() {
		return fmt.Errorf("Invalid WS event type")
	}

	active, err := s.loadModule(wsEvent.InstanceId)
	if err != nil {
		fmt.Printf("error loading module: %v\n", err)
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

	// Add timeout (defaults to 5 seconds)
	ctx, cancel := context.WithTimeout(ctx, s.maxExecutionTime)
	defer cancel()

	// this represents the number of current requests being processed
	// the Done() method removes 1 from the waitgroup. necessary for concurrency
	active.wg.Add(1)
	defer active.wg.Done()

	// This call is blocking, so we can ensure that once it returns we are complete
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
	s.mu.RLock()
	active, exists := s.activeModules[moduleId]
	s.mu.RUnlock()

	if exists {
		return active, nil
	}

	s.loadingModulesMu.Lock()
	signal, exists := s.loadingModules[moduleId]

	// The chan which signifies that this module is being loaded already exists, some other process is doing it
	if exists {
		fmt.Printf("Concurrent fetches: waiting on %s\n", moduleId)
		s.loadingModulesMu.Unlock()

		// wait on the signal
		// the max waiting time will be 5 seconds because of the time limit on fetching
		<-signal

		// if the module was loaded properly, this call will succeed and return the module
		// if not, it will initiate the loading process itself
		return s.loadModule(moduleId)
	}

	s.loadingModules[moduleId] = make(chan struct{})
	s.loadingModulesMu.Unlock()

	// signal that we are done loading if the function either returns or fails
	defer func() {
		s.loadingModulesMu.Lock()
		defer s.loadingModulesMu.Unlock()

		if ch, ok := s.loadingModules[moduleId]; ok {
			close(ch)
			delete(s.loadingModules, moduleId)
		}
	}()

	// Give the loader a 5 second timeout
	// the cancel function is deferred to avoid "context leaks"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	module, err := loader.Load(ctx, s.runtime, moduleId)
	if err != nil {
		return nil, err
	}

	mod := &ActiveModule{
		module:     module,
		lastUsed:   time.Now(),
		instanceId: moduleId,
	}

	// Add module to map of active modules
	// remove least recently used if we hit the limit
	s.mu.Lock()
	if len(s.activeModules) >= int(s.maxActiveModules) {
		s.evictLRU()
	}
	s.activeModules[moduleId] = mod
	s.mu.Unlock()

	return mod, nil
}
