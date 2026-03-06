package hostbuilder

import (
	"context"
	"fmt"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func fetchHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, urlPtr uint32, urlLen uint32, methodPtr uint32, methodLen uint32, bodyPtr uint32, bodyLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			logging.Logger.Errorf("Module memory is nil in fetchHandler")
			return 0
		}

		bytes, ok := mem.Read(urlPtr, urlLen)
		if !ok {
			logging.Logger.Errorf("Failed to read URL from module memory in fetchHandler")
			return 0
		}
		url := string(bytes)

		bytes, ok = mem.Read(methodPtr, methodLen)
		if !ok {
			logging.Logger.Errorf("Failed to read method from module memory in fetchHandler")
			return 0
		}
		method := string(bytes)

		bytes, ok = mem.Read(bodyPtr, bodyLen)
		if !ok {
			logging.Logger.Errorf("Failed to read body from module memory in fetchHandler")
			return 0
		}
		body := string(bytes)

		event, err := getWASMEvent(ctx, wasmevents.FETCH, url, method, body)
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in fetchHandler: %v", err)
			return 0
		}

		modCtx := getModuleContext(ctx, mod)

		resp, err := handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to execute FETCH handler: %v", err)
			ptr, _, _ := asmscript.CreateASError(
				modCtx,
				fmt.Errorf("Failed to execute FETCH hander"),
			)
			return uint32(ptr)
		}

		ptr, _, err := asmscript.CreateASString(
			modCtx,
			resp,
		)
		if err != nil {
			logging.Logger.Errorf("Failed to create string in WASM memory in fetchHandler: %v", err)
			ptr, _, _ := asmscript.CreateASError(
				modCtx,
				fmt.Errorf("Failed to execute FETCH hander"),
			)
			return uint32(ptr)
		}
		return uint32(ptr)
	}
}
