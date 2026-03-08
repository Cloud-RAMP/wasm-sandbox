package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func getUsersHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module) uint32 {
		event, err := getWASMEvent(ctx, wasmevents.GET_USERS, "")
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", wasmevents.GET_USERS.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), GET_WASM_EVENT_ERR)
		}

		modCtx := getModuleContext(ctx, mod)
		resp, err := handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to call handler in %s: %v", wasmevents.GET_USERS.String(), err)
			return writeErrorMessage(getModuleContext(ctx, mod), EXTERNAL_HANDLER_ERR)
		}

		ptr, _, err := asmscript.CreateASString(modCtx, resp)
		if err != nil {
			logging.Logger.Errorf("Failed to create string in WASM memory in getUsersHandler: %v", err)
			return writeErrorMessage(modCtx, CREATE_AS_STRING_ERR)
		}

		return uint32(ptr)
	}
}
