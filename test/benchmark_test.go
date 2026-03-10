package test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/store"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

const NUM_MODULES = 10

func dummyHandler(event *wasmevents.WASMEventInfo) (string, error) {
	return "dummy", nil
}

var abortChan chan string

func abortHandler(event *wasmevents.WASMEventInfo) (string, error) {
	abortChan <- event.Payload[0]
	return "", nil
}

func createNEvents(tb testing.TB, defaultEvent wsevents.WSEventInfo, n int) []*wsevents.WSEventInfo {
	tb.Helper()
	out := make([]*wsevents.WSEventInfo, 0)

	for i := range n {
		event := defaultEvent
		event.InstanceId = fmt.Sprintf("./modules/%d.wasm", i+1)
		out = append(out, &event)
	}

	return out
}

func setupStore(tb testing.TB) *store.SandboxStore {
	tb.Helper()
	store, err := store.NewSandboxStore(store.SandboxStoreCfg{
		CleanupInterval:    5 * time.Second,
		MaxIdleTime:        6 * time.Second,
		MemoryLimitPages:   10,
		MaxActiveModules:   NUM_MODULES - 1,
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
	ctx := b.Context()

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
	ctx, cancel := context.WithCancel(b.Context())
	defer cancel()

	event := wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "./modules/1.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}
	events := createNEvents(b, event, NUM_MODULES)

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

func BenchmarkZipfWithModuleEviction(b *testing.B) {
	store := setupStore(b)
	ctx, cancel := context.WithCancel(b.Context())
	defer cancel()

	event := wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "./modules/1.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}
	events := createNEvents(b, event, NUM_MODULES)

	abortChan = make(chan string)
	go func() {
		for msg := range abortChan {
			b.Logf("Abort called: %s\n", msg)
			cancel()
			return
		}
	}()

	// use zipf for typical service access patterns
	r := rand.New(rand.NewSource(42))
	zipf := rand.NewZipf(r, 1.5, 1, uint64(len(events)-1))

	for b.Loop() {
		select {
		case <-ctx.Done():
			return
		default:
			i := zipf.Uint64()
			err := store.ExecuteOnModule(ctx, events[i])
			if err != nil {
				b.Fatalf("Failed to execute on module %s: %v\n", events[i].InstanceId, err)
			}
		}
	}

	close(abortChan)
}

func BenchmarkZipfWithoutModuleEviction(b *testing.B) {
	store := setupStore(b)
	ctx, cancel := context.WithCancel(b.Context())
	defer cancel()

	event := wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "./modules/1.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}
	events := createNEvents(b, event, NUM_MODULES-1)

	abortChan = make(chan string)
	go func() {
		for msg := range abortChan {
			b.Logf("Abort called: %s\n", msg)
			cancel()
			return
		}
	}()

	// use zipf for typical service access patterns
	r := rand.New(rand.NewSource(5980))
	zipf := rand.NewZipf(r, 1.5, 1, uint64(len(events)-1))

	for b.Loop() {
		select {
		case <-ctx.Done():
			return
		default:
			i := zipf.Uint64()
			err := store.ExecuteOnModule(ctx, events[i])
			if err != nil {
				b.Fatalf("Failed to execute on module %s: %v\n", events[i].InstanceId, err)
			}
		}
	}

	close(abortChan)
}
