package test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/pkg/store"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

var (
	setupOnce  sync.Once
	benchStore *store.SandboxStore
	benchCtx   context.Context
	benchEvent *wsevents.WSEventInfo
)

func BenchmarkParallelSingleModule(b *testing.B) {
	// setup text
	setupOnce.Do(func() {
		benchStore = setupStore(b)
		benchCtx = context.Background()

		benchEvent = &wsevents.WSEventInfo{
			ConnectionId: "bench-connection",
			InstanceId:   "../example/build/release.wasm",
			RoomId:       "bench-room",
			Payload:      "benchmark payload",
			EventType:    wsevents.ON_MESSAGE,
			Timestamp:    time.Now().UnixMilli(),
		}

		// Pre-warm the module
		err := benchStore.ExecuteOnModule(benchCtx, benchEvent)
		if err != nil {
			b.Fatalf("Failed to execute module: %v", err)
		}
		time.Sleep(1 * time.Second)
	})

	b.SetParallelism(8)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := benchStore.ExecuteOnModule(benchCtx, benchEvent)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkParallelModuleEviction(b *testing.B) {
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

	events := make([]*wsevents.WSEventInfo, 0, NUM_MODULES)
	events = append(events, &event)
	for i := range NUM_MODULES - 1 {
		e := event
		e.InstanceId = fmt.Sprintf("./modules/%d.wasm", i+2)
		events = append(events, &e)
	}

	abortChan := make(chan string)
	go func() {
		for msg := range abortChan {
			b.Logf("Abort called: %s\n", msg)
			cancel()
			return
		}
	}()

	eventsLen := int64(len(events))
	var counter atomic.Int64

	b.SetParallelism(8)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			select {
			case <-ctx.Done():
				return
			default:
				i := counter.Add(1) - 1
				ev := events[i%eventsLen]
				if err := store.ExecuteOnModule(ctx, ev); err != nil {
					b.Errorf("Failed to execute on module %s: %v\n", ev.InstanceId, err)
					cancel()
					return
				}
			}
		}
	})

	close(abortChan)
}

func BenchmarkMultipleModulesParallel(b *testing.B) {
	store := setupStore(b)
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	event := wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "./modules/1.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}

	events := make([]*wsevents.WSEventInfo, 0, NUM_MODULES)
	events = append(events, &event)
	for i := range NUM_MODULES - 1 {
		e := event
		e.InstanceId = fmt.Sprintf("./modules/%d.wasm", i+2)
		events = append(events, &e)
	}

	abortChan := make(chan string) // local, not package-level
	go func() {
		for msg := range abortChan {
			b.Logf("Abort called: %s\n", msg)
			cancel()
			return
		}
	}()

	eventsLen := int64(len(events))
	var counter atomic.Int64

	b.SetParallelism(8)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			select {
			case <-ctx.Done():
				return
			default:
				i := counter.Add(1) - 1
				ev := events[i%eventsLen]
				if err := store.ExecuteOnModule(ctx, ev); err != nil {
					b.Errorf("Failed to execute on module %s: %v\n", ev.InstanceId, err)
					// cancel()
					return
				}
			}
		}
	})

	close(abortChan)
}
