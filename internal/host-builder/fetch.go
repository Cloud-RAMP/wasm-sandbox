package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func fetchHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, urlPtr uint32, urlLen uint32, methodPtr uint32, methodLen uint32, bodyPtr uint32, bodyLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), MOD_MEMORY_ERR)
		}

		bytes, ok := mem.Read(urlPtr, urlLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		url := string(bytes)

		bytes, ok = mem.Read(methodPtr, methodLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		method := string(bytes)

		bytes, ok = mem.Read(bodyPtr, bodyLen)
		if !ok {
			return writeErrorMessage(getModuleContext(ctx, mod), MEM_READ_ERR)
		}
		body := string(bytes)

		event, err := getWASMEvent(ctx, wasmevents.FETCH, url, method, body)
		if event == nil {
			return writeErrorMessage(getModuleContext(ctx, mod), GET_WASM_EVENT_ERR)
		}

		modCtx := getModuleContext(ctx, mod)

		resp, err := handlerMap.CallHandler(event)
		if err != nil {
			return writeErrorMessage(getModuleContext(ctx, mod), EXTERNAL_HANDLER_ERR)
		}

		ptr, _, err := asmscript.CreateASString(
			modCtx,
			resp,
		)
		if err != nil {
			return writeErrorMessage(modCtx, CREATE_AS_STRING_ERR)
		}
		return uint32(ptr)
	}
}
