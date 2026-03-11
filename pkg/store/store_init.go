package store

import (
	"context"
	"time"

	builder "github.com/Cloud-RAMP/wasm-sandbox/internal/host-builder"
	"github.com/tetratelabs/wazero"
)

// Genertic function to get a default value. Done to reduce boilerplate if statements
func defaultValue[T comparable](original T, zeroCase T, defaultValue T) T {
	if original == zeroCase {
		return defaultValue
	}
	return original
}

// Will probably need to pass a ctx into this later, or limit execution time somehow
func NewSandboxStore(cfg SandboxStoreCfg) (*SandboxStore, error) {
	ctx := context.Background()

	memPages := defaultValue(cfg.MemoryLimitPages, 0, 10)

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
		runtime:          runtime,
		hostModule:       hostModule,
		moduleConfig:     wazero.NewModuleConfig(),
		activeModules:    make(map[string]*ActiveModule),
		loadingModules:   make(map[string]chan struct{}),
		maxActiveModules: defaultValue(cfg.MaxActiveModules, 0, 25),
		maxExecutionTime: defaultValue(cfg.MaxExecutionTime, 0, 5*time.Second),
	}

	// auto-clean up modules if cleanup interval and max idle time are defined
	if cfg.CleanupInterval != 0 && cfg.MaxIdleTime != 0 {
		store.cleanupInterval = cfg.CleanupInterval
		store.maxIdleTime = cfg.MaxIdleTime
		store.startCleanupRoutine()
	}

	return store, nil
}
