package builder

import (
	"context"
	"fmt"
	"log"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// TODO: imrove the process of adding more functions to a module
func BuildHostModule(runtime wazero.Runtime) (api.Module, error) {
	hostModuleBuilder := runtime.NewHostModuleBuilder("env")

	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, messagePtr uint32, fileNamePtr uint32, line uint32, column uint32) {
			if mod != nil {
				message := asmscript.ReadASString(mod.Memory(), messagePtr)
				fileName := asmscript.ReadASString(mod.Memory(), fileNamePtr)
				log.Printf("AssemblyScript abort: %s at %s:%d:%d", message, fileName, line, column)
			}
		}).
		Export("abort")

	// Broadcast function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, ptr uint32, len uint32) {
			mem := mod.Memory()
			if mem == nil {
				return
			}

			bytes, ok := mem.Read(ptr, len)
			if !ok {
				return
			}
			message := string(bytes)

			// Need to get the service instance to store events
			// This is tricky - we'll address this below
			fmt.Printf("Message received: %s\n", message)
		}).
		Export("broadcast")

	return hostModuleBuilder.Instantiate(context.Background())
}
