package hostbuilder

import (
	"context"
	"fmt"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func getHandler(handlerMap *wasmevents.HandlerMap, getType wasmevents.WASMEventType) any {
	return func(ctx context.Context, mod api.Module, keyPtr uint32, keyLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return 0
		}

		bytes, ok := mem.Read(keyPtr, keyLen)
		if !ok {
			return 0
		}

		event, err := getWASMEvent(ctx, getType, string(bytes))
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", getType.String(), err)
			return 0
		}

		modCtx := getModuleContext(ctx, mod)
		val, err := handlerMap.CallHandler(event)
		if err != nil {
			ptr, _, _ := asmscript.CreateASError(
				modCtx,
				fmt.Errorf("Failed to execute handler %s: %v", getType, err),
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
