package hostbuilder

import (
	"context"
	"fmt"

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
			return 0
		}

		modCtx, err := getModuleContext(ctx, mod)
		if err != nil {
			logging.Logger.Errorf("Failed to getModuleContext in handler %s: %v", wasmevents.GET_USERS.String(), err)
			return 0
		}

		resp, err := handlerMap.CallHandler(event)
		if err != nil {
			ptr, _, _ := asmscript.CreateASError(
				modCtx,
				fmt.Errorf("Failed to call handler in %s: %v", wasmevents.GET_USERS.String(), err),
			)
			return uint32(ptr)
		}

		ptr, _, err := asmscript.CreateASString(modCtx, resp)
		if err != nil {
			ptr, _, _ := asmscript.CreateASError(
				modCtx,
				fmt.Errorf("Failed to call handler in %s: %v", wasmevents.GET_USERS.String(), err),
			)
			return uint32(ptr)
		}

		return uint32(ptr)
	}
}
