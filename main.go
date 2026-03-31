package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/pkg/loader"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/store"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

func dummyHandler(event *wasmevents.WASMEventInfo) (string, error) {
	// fmt.Println("New event")
	// fmt.Printf("Event %s from %s\n", event.EventType.String(), event.InstanceId)
	// fmt.Println("Args:", event.Payload)
	return "dummy", nil
}

func debugHandler(event *wasmevents.WASMEventInfo) (string, error) {
	fmt.Printf("WASM DEBUG: %v\n", event.Payload)
	return "", nil
}

// make the main sandbox functions that we expose here
func main() {
	ctx := context.Background()

	store, err := store.NewSandboxStore(context.Background(), store.SandboxStoreCfg{
		CleanupInterval:    5 * time.Second,
		MaxIdleTime:        6 * time.Second,
		MemoryLimitPages:   10,
		CloseOnContextDone: true,
		HandlerMap: wasmevents.NewHandlerMap().
			AddHandler(wasmevents.GET, dummyHandler).
			AddHandler(wasmevents.SET, dummyHandler).
			AddHandler(wasmevents.DEL, dummyHandler).
			AddHandler(wasmevents.DB_GET, dummyHandler).
			AddHandler(wasmevents.DB_SET, dummyHandler).
			AddHandler(wasmevents.DB_DEL, dummyHandler).
			AddHandler(wasmevents.BROADCAST, dummyHandler).
			AddHandler(wasmevents.FETCH, dummyHandler).
			AddHandler(wasmevents.LOG, dummyHandler).
			AddHandler(wasmevents.DEBUG, debugHandler).
			AddHandler(wasmevents.GET_USERS, dummyHandler).
			AddHandler(wasmevents.SEND_MESSAGE, dummyHandler),
		LoaderFunction: loader.MockLoaderFunction,
	})

	if err != nil {
		fmt.Println("Failed to make sandbox store", err)
		return
	}

	var wg sync.WaitGroup

	for i := range 10 {
		// sample event
		event := &wsevents.WSEventInfo{
			ConnectionId: fmt.Sprintf("connection-%d", i),
			InstanceId:   "example/build/release.wasm", // simple loader function is only configued to use filenames as instance IDs
			RoomId:       "first-room",
			Payload:      "hello, world!",
			EventType:    wsevents.ON_MESSAGE,
			Timestamp:    time.Now().UnixMilli(),
		}

		wg.Add(1)

		go func() {
			store.ExecuteOnModule(ctx, event)
			wg.Done()
		}()
	}

	fmt.Println("Finished sending requests")
	wg.Wait()
}
