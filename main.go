package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/store"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

func dummyHandler(event wasmevents.WASMEventInfo) (string, error) {
	fmt.Println("New event")
	fmt.Printf("Event %s from %s\n", event.EventType.String(), event.InstanceId)
	fmt.Println("Args:", event.Payload)
	return "dummy", nil
}

func debugHandler(event wasmevents.WASMEventInfo) (string, error) {
	fmt.Printf("WASM DEBUG: %v\n", event.Payload)
	return "", nil
}

// make the main sandbox functions that we expose here
func main() {
	ctx := context.Background()

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
			AddHandler(wasmevents.DEBUG, debugHandler).
			AddHandler(wasmevents.GET_USERS, dummyHandler).
			AddHandler(wasmevents.SEND_MESSAGE, dummyHandler),
	})

	if err != nil {
		fmt.Println("Failed to make sandbox store", err)
		return
	}

	// sample event
	event := wsevents.WSEventInfo{
		ConnectionId: "first-connection",
		InstanceId:   "example/build/release.wasm", // simple loader function is only configued to use filenames as instance IDs
		RoomId:       "first-room",
		Payload:      "hello, world!",
		EventType:    wsevents.ON_MESSAGE,
		Timestamp:    time.Now().UnixMilli(),
	}

	go store.ExecuteOnModule(ctx, event)

	// // second event to introduce concurrency issues
	// event = wsevents.WSEventInfo{
	// 	ConnectionId: "second-connection",
	// 	InstanceId:   "example/build/release.wasm",
	// 	RoomId:       "first-room",
	// 	Payload:      "hello, world!",
	// 	EventType:    wsevents.ON_MESSAGE,
	// 	Timestamp:    time.Now().UnixMilli(),
	// }
	// go store.ExecuteOnModule(ctx, event)

	// sleep so that all events can be read
	// won't need this in the server, as it will be a long running process
	time.Sleep(3 * time.Second)
}
