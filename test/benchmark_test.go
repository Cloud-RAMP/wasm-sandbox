package test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

// BenchmarkSimpleSingleModule-8              78194             13373 ns/op           24783 B/op        39 allocs/op
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
	events := make([]*wsevents.WSEventInfo, 0)
	events = append(events, &event)
	for i := range NUM_MODULES - 1 {
		e := event
		e.InstanceId = fmt.Sprintf("./modules/%d.wasm", i+2)
		events = append(events, &e)
	}

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

func BenchmarkMultipleModulesNoEviction(b *testing.B) {
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
	events := make([]*wsevents.WSEventInfo, 0)
	events = append(events, &event)
	for i := range NUM_MODULES - 1 {
		e := event
		e.InstanceId = fmt.Sprintf("./modules/%d.wasm", i+2)
		events = append(events, &e)
	}

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

func BenchmarkZipf(b *testing.B) {
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
	events := make([]*wsevents.WSEventInfo, 0)
	events = append(events, &event)
	for i := range NUM_MODULES - 1 {
		e := event
		e.InstanceId = fmt.Sprintf("./modules/%d.wasm", i+2)
		events = append(events, &e)
	}

	abortChan = make(chan string)
	go func() {
		for msg := range abortChan {
			b.Logf("Abort called: %s\n", msg)
			cancel()
			return
		}
	}()

	rng := rand.New(rand.NewSource(42)) // deterministic benchmark distribution
	zipf := rand.NewZipf(rng, 1.2, 1, uint64(len(events)-1))

	// Testing loop
	for b.Loop() {
		select {
		case <-ctx.Done():
			return
		default:
			idx := int(zipf.Uint64())
			err := store.ExecuteOnModule(ctx, events[idx])
			if err != nil {
				b.Fatalf("Failed to execute on module %s: %v\n", events[idx].InstanceId, err)
			}
		}
	}

	close(abortChan)
}
