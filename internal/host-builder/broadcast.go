package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func broadcastHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, ptr uint32, len uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(ptr, len)
		if !ok {
			return
		}

		event, err := getWASMEvent(ctx, wasmevents.BROADCAST, string(bytes))
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", wasmevents.BROADCAST.String(), err)
			return
		}

		_, err = handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to execute handler %s: %v", wasmevents.BROADCAST.String(), err)
		}
	}
}
