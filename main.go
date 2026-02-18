package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/store"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/events"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/handlers"
)

func dummyHandler(event events.EventType, instanceId string, args ...string) (string, error) {
	fmt.Println("New event")
	fmt.Printf("Event %s from %s\n", event.String(), instanceId)
	fmt.Println("Args:", args)
	return "dummy", nil
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
		HandlerMap: events.NewHandlerMap().
			AddHandler(events.GET, dummyHandler).
			AddHandler(events.SET, dummyHandler).
			AddHandler(events.BROADCAST, dummyHandler),
	})

	if err != nil {
		fmt.Println("Failed to make sandbox store", err)
		return
	}

	if err := store.LoadModule("first", wasmBytes); err != nil {
		fmt.Println("Failed to load module", err)
		return
	}

	go store.ExecuteOnModule(ctx, "first", handlers.ON_MESSAGE, "hello, world!")

	// sleep so that all events can be read
	// won't need this in the server, as it will be a long running process
	time.Sleep(3 * time.Second)
}
