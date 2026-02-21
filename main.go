package main

import (
	"context"
	"fmt"
	"os"
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

	wasmBytes, err := os.ReadFile("./example/build/release.wasm")
	if err != nil {
		fmt.Println("Failed to read wasm file", err)
		return
	}

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
			AddHandler(wasmevents.DEBUG, debugHandler),
	})

	if err != nil {
		fmt.Println("Failed to make sandbox store", err)
		return
	}

	if err := store.LoadModuleIntoSandbox("first-instance", wasmBytes); err != nil {
		fmt.Println("Failed to load module", err)
		return
	}

	// sample event
	event := wsevents.WSEventInfo{
		ConnectionId: "first-connection",
		InstanceId:   "first-instance",
		RoomId:       "first-room",
		Payload:      "hello, world!",
		EventType:    wsevents.ON_JOIN,
		Timestamp:    time.Now().UnixMilli(),
	}

	go store.ExecuteOnModule(ctx, event)

	// sleep so that all events can be read
	// won't need this in the server, as it will be a long running process
	time.Sleep(3 * time.Second)
}
