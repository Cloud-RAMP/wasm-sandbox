package hostbuilder

import (
	"context"
	"fmt"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func getHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, keyPtr uint32, keyLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return 0
		}

		bytes, ok := mem.Read(keyPtr, keyLen)
		if !ok {
			return 0
		}

		event, err := getWASMEvent(ctx, wasmevents.GET, string(bytes))
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", wasmevents.GET.String(), err)
			return 0
		}

		modCtx, err := getModuleContext(ctx, mod)
		if err != nil {
			logging.Logger.Errorf("Failed to getModuleContext in handler %s: %v", wasmevents.FETCH.String(), err)
			return 0
		}

		val, err := handlerMap.CallHandler(event)
		if err != nil {
			ptr, _, _ := asmscript.CreateASError(
				modCtx,
				fmt.Errorf("Failed to execute GET hander"),
			)
			return uint32(ptr)
		}

		ptr, _, err := asmscript.CreateASString(
			modCtx,
			val,
		)
		if err != nil {
			return 0
		}

		return uint32(ptr)
	}
}
