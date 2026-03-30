package hostbuilder

import (
	"context"
	"strings"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func getUsersHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module) uint32 {
		event, err := getWASMEvent(ctx, wasmevents.GET_USERS, "")
		if event == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), GET_WASM_EVENT_ERR)
		}

		modCtx := getModuleContext(ctx, mod)
		resp, err := handlerMap.CallHandler(event)
		if err != nil {
			return writeErrorMessage(modCtx, EXTERNAL_HANDLER_ERR)
		}

		respSplit := strings.Split(resp, ",")
		ptr, _, err := asmscript.WriteArray(modCtx, respSplit)
		if err != nil {
			return writeErrorMessage(modCtx, CREATE_AS_STRING_ERR)
		}

		return uint32(ptr)
	}
}
