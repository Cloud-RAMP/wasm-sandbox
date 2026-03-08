package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func sendMessageHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, userPtr uint32, userLen uint32, msgPtr uint32, msgLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), MOD_MEMORY_ERR)
		}

		bytes, ok := mem.Read(userPtr, userLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		user := string(bytes)

		bytes, ok = mem.Read(msgPtr, msgLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		msg := string(bytes)

		event, err := getWASMEvent(ctx, wasmevents.SEND_MESSAGE, user, msg)
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", wasmevents.SEND_MESSAGE.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), GET_WASM_EVENT_ERR)
		}

		_, err = handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to call handler in %s: %v", wasmevents.SEND_MESSAGE.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), EXTERNAL_HANDLER_ERR)
		}

		return 0
	}
}
