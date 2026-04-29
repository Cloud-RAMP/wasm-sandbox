package test

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

func BenchmarkParallelSingleModule(b *testing.B) {
	// setup text
	benchStore := setupStore(b, 1)
	benchCtx := context.Background()

	benchEvent := &wsevents.WSEventInfo{
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
	store := setupStore(b, 5)
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
}

func BenchmarkMultipleModulesParallel(b *testing.B) {
	store := setupStore(b, 5)
	ctx := context.Background()

	event := wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "./modules/1.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}

	events := make([]*wsevents.WSEventInfo, 0, 5)
	events = append(events, &event)
	for i := range 5 - 1 {
		e := event
		e.InstanceId = fmt.Sprintf("./modules/%d.wasm", i+2)
		events = append(events, &e)
	}

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
					b.Fatalf("Failed to execute on module %s: %v\n", ev.InstanceId, err)
					// cancel()
					return
				}
			}
		}
	})
}

func BenchmarkParallelZipf(b *testing.B) {
	const STORE_MODULES = 5
	store := setupStore(b, STORE_MODULES)
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
	events := make([]*wsevents.WSEventInfo, 0, STORE_MODULES)
	events = append(events, &event)
	for i := range STORE_MODULES - 1 {
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

	// each goroutine gets its own rng and zipf to avoid shared state
	b.SetParallelism(8)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		zipf := rand.NewZipf(rng, 1.2, 1, uint64(len(events)-1))

		for pb.Next() {
			select {
			case <-ctx.Done():
				return
			default:
				idx := int(zipf.Uint64())
				ev := events[idx]
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

func BenchmarkLatencyPercentiles(b *testing.B) {
	store := setupStore(b, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := &wsevents.WSEventInfo{
		ConnectionId: "bench-connection",
		InstanceId:   "./modules/1.wasm",
		RoomId:       "bench-room",
		Payload:      "benchmark payload",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}

	// warmup
	for range 10 {
		if err := store.ExecuteOnModule(ctx, event); err != nil {
			b.Fatalf("Warmup failed: %v", err)
		}
	}

	latencies := make([]time.Duration, 0, b.N)
	var mu sync.Mutex

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			start := time.Now()
			err := store.ExecuteOnModule(ctx, event)
			elapsed := time.Since(start)

			if err != nil {
				b.Errorf("Failed to execute: %v", err)
				cancel()
				return
			}

			mu.Lock()
			latencies = append(latencies, elapsed)
			mu.Unlock()
		}
	})

	b.StopTimer()

	// sort and compute percentiles
	slices.Sort(latencies)

	n := len(latencies)
	if n == 0 {
		b.Fatal("No latencies recorded")
	}

	p50 := latencies[n*50/100]
	p95 := latencies[n*95/100]
	p99 := latencies[n*99/100]
	pMax := latencies[n-1]

	b.ReportMetric(float64(p50.Microseconds()), "p50_us")
	b.ReportMetric(float64(p95.Microseconds()), "p95_us")
	b.ReportMetric(float64(p99.Microseconds()), "p99_us")
	b.ReportMetric(float64(pMax.Microseconds()), "pMax_us")

	// also print for readability
	b.Logf("p50:  %v", p50)
	b.Logf("p95:  %v", p95)
	b.Logf("p99:  %v", p99)
	b.Logf("pMax: %v", pMax)
}
