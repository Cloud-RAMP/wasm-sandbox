package store

import (
	"context"
	"time"

	builder "github.com/Cloud-RAMP/wasm-sandbox/internal/host-builder"
	"github.com/tetratelabs/wazero"
)

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

	maxActiveModules := cfg.MaxActicveModules
	if maxActiveModules == 0 {
		maxActiveModules = 25
	}

	store := &SandboxStore{
		runtime:          runtime,
		hostModule:       hostModule,
		moduleConfig:     wazero.NewModuleConfig(),
		activeModules:    make(map[string]*ActiveModule),
		maxActiveModules: maxActiveModules,
	}

	// auto-clean up modules if cleanup interval and max idle time are defined
	if cfg.CleanupInterval != 0 && cfg.MaxIdleTime != 0 {
		store.cleanupInterval = cfg.CleanupInterval
		store.maxIdleTime = cfg.MaxIdleTime
		store.startCleanupRoutine()
	}

	// default to 5 seconds
	if cfg.MaxExecutionTime == 0 {
		cfg.MaxExecutionTime = 5 * time.Second
	}

	return store, nil
}
