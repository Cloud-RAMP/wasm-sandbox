package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func setHandler(handlerMap *wasmevents.HandlerMap, setType wasmevents.WASMEventType) any {
	return func(ctx context.Context, mod api.Module, keyPtr uint32, keyLen uint32, valPtr uint32, valLen uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(keyPtr, keyLen)
		if !ok {
			return
		}
		key := string(bytes)

		bytes, ok = mem.Read(valPtr, valLen)
		if !ok {
			return
		}
		val := string(bytes)

		event, err := getWASMEvent(ctx, setType, key, val)
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", setType.String(), err)
			return
		}

		_, err = handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to execute handler %s: %v", setType.String(), err)
		}
	}
}
