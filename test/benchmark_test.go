package test

import (
	"context"
	"testing"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/store"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

func dummyHandler(event *wasmevents.WASMEventInfo) (string, error) {
	return "dummy", nil
}

var abortChan chan string

func abortHandler(event *wasmevents.WASMEventInfo) (string, error) {
	abortChan <- event.Payload[0]
	return "", nil
}

func setupStore(tb testing.TB) *store.SandboxStore {
	tb.Helper()
	store, err := store.NewSandboxStore(store.SandboxStoreCfg{
		CleanupInterval:    5 * time.Second,
		MaxIdleTime:        6 * time.Second,
		MemoryLimitPages:   10,
		MaxActiveModules:   3,
		CloseOnContextDone: true,
		HandlerMap: wasmevents.NewHandlerMap().
			AddHandler(wasmevents.ABORT, abortHandler).
			AddHandler(wasmevents.GET, dummyHandler).
			AddHandler(wasmevents.SET, dummyHandler).
			AddHandler(wasmevents.DB_GET, dummyHandler).
			AddHandler(wasmevents.DB_SET, dummyHandler).
			AddHandler(wasmevents.BROADCAST, dummyHandler).
			AddHandler(wasmevents.LOG, dummyHandler).
			AddHandler(wasmevents.DEBUG, dummyHandler).
			AddHandler(wasmevents.GET_USERS, dummyHandler).
			AddHandler(wasmevents.SEND_MESSAGE, dummyHandler).
			AddHandler(wasmevents.FETCH, dummyHandler),
	})
	if err != nil {
		tb.Fatalf("Failed to make sandbox store: %v", err)
	}
	return store
}

func BenchmarkSimpleSingleModule(b *testing.B) {
	store := setupStore(b)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := &wsevents.WSEventInfo{
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

func BenchmarkSimpleModuleEviction(b *testing.B) {
	store := setupStore(b)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "./modules/1.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}
	event2 := event
	event2.InstanceId = "./modules/2.wasm"
	event3 := event
	event3.InstanceId = "./modules/3.wasm"
	event4 := event
	event4.InstanceId = "./modules/4.wasm"

	events := []*wsevents.WSEventInfo{&event, &event2, &event3, &event4}

	abortChan = make(chan string)
	go func() {
		for msg := range abortChan {
			b.Logf("Abort called: %s\n", msg)
			cancel()
			return
		}
	}()

	// Testing loop
	i := 0
	eventsLen := len(events)
	for b.Loop() {
		select {
		case <-ctx.Done():
			return
		default:
			err := store.ExecuteOnModule(ctx, events[i%eventsLen])
			if err != nil {
				b.Fatalf("Failed to execute on module %s: %v\n", events[i%eventsLen].InstanceId, err)
			}
			i++
		}
	}

	close(abortChan)
}
