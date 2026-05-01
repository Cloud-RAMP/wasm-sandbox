package test

import (
	"testing"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/pkg/loader"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/store"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
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

func setupStore(tb testing.TB, maxModules int) *store.SandboxStore {
	tb.Helper()
	store, err := store.NewSandboxStore(tb.Context(), store.SandboxStoreCfg{
		CleanupInterval:    5 * time.Second,
		MaxIdleTime:        6 * time.Second,
		MemoryLimitPages:   10,
		MaxActiveModules:   uint16(maxModules),
		CloseOnContextDone: true,
		PoolSize:           uint8(10),
		HandlerMap: wasmevents.NewHandlerMap().
			AddHandler(wasmevents.ABORT, abortHandler).
			AddHandler(wasmevents.GET, dummyHandler).
			AddHandler(wasmevents.SET, dummyHandler).
			AddHandler(wasmevents.DEL, dummyHandler).
			AddHandler(wasmevents.DB_GET, dummyHandler).
			AddHandler(wasmevents.DB_SET, dummyHandler).
			AddHandler(wasmevents.DB_DEL, dummyHandler).
			AddHandler(wasmevents.BROADCAST, dummyHandler).
			AddHandler(wasmevents.LOG, dummyHandler).
			AddHandler(wasmevents.DEBUG, dummyHandler).
			AddHandler(wasmevents.GET_USERS, dummyHandler).
			AddHandler(wasmevents.SEND_MESSAGE, dummyHandler).
			AddHandler(wasmevents.SERVER_MESSAGE, dummyHandler).
			AddHandler(wasmevents.CLOSE_CONNECTION, dummyHandler).
			AddHandler(wasmevents.FETCH, dummyHandler),
		LoaderFunction: loader.MockLoaderFunction,
	})
	if err != nil {
		tb.Fatalf("Failed to make sandbox store: %v", err)
	}
	return store
}
