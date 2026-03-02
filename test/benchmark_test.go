package test

import (
	"context"
	"testing"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/store"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

func dummyHandler(event wasmevents.WASMEventInfo) (string, error) {
	return "dummy", nil
}

func setupStore(tb testing.TB) *store.SandboxStore {
	tb.Helper()
	store, err := store.NewSandboxStore(store.SandboxStoreCfg{
		CleanupInterval:    5 * time.Second,
		MaxIdleTime:        6 * time.Second,
		MemoryLimitPages:   10,
		CloseOnContextDone: true,
		HandlerMap: wasmevents.NewHandlerMap().
			AddHandler(wasmevents.GET, dummyHandler).
			AddHandler(wasmevents.SET, dummyHandler).
			AddHandler(wasmevents.BROADCAST, dummyHandler).
			AddHandler(wasmevents.LOG, dummyHandler).
			AddHandler(wasmevents.DEBUG, dummyHandler).
			AddHandler(wasmevents.GET_USERS, dummyHandler).
			AddHandler(wasmevents.SEND_MESSAGE, dummyHandler),
	})
	if err != nil {
		tb.Fatalf("Failed to make sandbox store: %v", err)
	}
	return store
}

func BenchmarkExecuteOnModule(b *testing.B) {
	store := setupStore(b)
	ctx := context.Background()

	event := wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "../example/build/release.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}

	// Pre-warm the module
	err := store.ExecuteOnModule(ctx, event)
	if err != nil {
		b.Fatalf("Failed to execute module: %v", err)
	}
	time.Sleep(1 * time.Second)
	b.ResetTimer()

	// Testing loop
	for b.Loop() {
		_ = store.ExecuteOnModule(ctx, event)
	}
}
