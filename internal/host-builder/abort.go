package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

func abortHandler(handlers *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, messagePtr uint32, fileNamePtr uint32, line uint32, column uint32) {
		if mod != nil {
			message := asmscript.ReadASString(mod.Memory(), messagePtr)
			// fileName := asmscript.ReadASString(mod.Memory(), fileNamePtr)
			// logging.Logger.Errorf("AssemblyScript abort: %s at %s:%d:%d", message, fileName, line, column)

			event, _ := getWASMEvent(ctx, wasmevents.ABORT, message)
			handlers.CallHandler(event)
			// os.Exit(1)
		}
	}
}
