package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func getHandler(handlerMap *wasmevents.HandlerMap, getType wasmevents.WASMEventType) any {
	return func(ctx context.Context, mod api.Module, keyPtr uint32, keyLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), MOD_MEMORY_ERR)
		}

		bytes, ok := mem.Read(keyPtr, keyLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}

		event, err := getWASMEvent(ctx, getType, string(bytes))
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", getType.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), GET_WASM_EVENT_ERR)
		}

		modCtx := getModuleContext(ctx, mod)
		val, err := handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to call handler in %s: %v", getType.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), EXTERNAL_HANDLER_ERR)
		}

		ptr, _, err := asmscript.CreateASString(
			modCtx,
			val,
		)
		if err != nil {
			logging.Logger.Errorf("Failed to create string in WASM memory in getHandler: %v", err)
			return writeErrorMessage(modCtx, CREATE_AS_STRING_ERR)
		}

		return uint32(ptr)
	}
}
