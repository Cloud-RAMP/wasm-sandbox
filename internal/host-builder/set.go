package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func setHandler(handlerMap *wasmevents.HandlerMap, setType wasmevents.WASMEventType) any {
	return func(ctx context.Context, mod api.Module, keyPtr uint32, keyLen uint32, valPtr uint32, valLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), MOD_MEMORY_ERR)
		}

		bytes, ok := mem.Read(keyPtr, keyLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		key := string(bytes)

		bytes, ok = mem.Read(valPtr, valLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		val := string(bytes)

		event, err := getWASMEvent(ctx, setType, key, val)
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", setType.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), GET_WASM_EVENT_ERR)
		}

		_, err = handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to execute handler %s: %v", setType.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), EXTERNAL_HANDLER_ERR)
		}

		return 0
	}
}
