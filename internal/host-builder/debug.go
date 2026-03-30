package hostbuilder

import (
	"context"

	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func debugHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, strPtr uint32, strLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), MOD_MEMORY_ERR)
		}

		bytes, ok := mem.Read(strPtr, strLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		info := string(bytes)

		event, err := getWASMEvent(ctx, wasmevents.DEBUG, info)
		if event == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), GET_WASM_EVENT_ERR)
		}
		_, err = handlerMap.CallHandler(event)
		if err != nil {
			return writeErrorMessage(getModuleContext(ctx, mod), EXTERNAL_HANDLER_ERR)
		}

		return 0
	}
}
