package hostbuilder

import (
	"context"
	"fmt"
	"log"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

// These functions all return closures that capture the value of events.HandlerMap becuause it avoids circular import issues

func abortHandler(_ *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, messagePtr uint32, fileNamePtr uint32, line uint32, column uint32) {
		if mod != nil {
			message := asmscript.ReadASString(mod.Memory(), messagePtr)
			fileName := asmscript.ReadASString(mod.Memory(), fileNamePtr)
			log.Printf("AssemblyScript abort: %s at %s:%d:%d", message, fileName, line, column)
		}
	}
}

func broadcastHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, ptr uint32, len uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(ptr, len)
		if !ok {
			return
		}

		message := string(bytes)
		instanceId := ctx.Value("instanceId").(string)

		_, _ = handlerMap.CallHandler(wasmevents.BROADCAST, instanceId, message)
	}
}

func setHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, keyPtr uint32, keyLen uint32, valPtr uint32, valLen uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(keyPtr, keyLen)
		if !ok {
			return
		}
		key := string(bytes)

		bytes, ok = mem.Read(valPtr, valLen)
		if !ok {
			return
		}
		val := string(bytes)

		fmt.Printf("SET request %s for %s\n", key, val)

		instanceId := ctx.Value("instanceId").(string)

		handlerMap.CallHandler(wasmevents.SET, instanceId, key, val)
	}
}

// Returns a uint32 becuase it is the location in wasm memory of the returned string
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
		key := string(bytes)
		instanceId := ctx.Value("instanceId").(string)

		// do some sort of redis operation here
		val, err := handlerMap.CallHandler(wasmevents.GET, instanceId, key)
		if err != nil {
			// some sort of error handling here
			return 0
		}

		// insert the string into module memory
		ptr, _, err := asmscript.CreateASString(mod, val)
		if err != nil {
			return 0
		}

		return uint32(ptr)
	}
}

func logHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, keyPtr uint32, keyLen uint32, valPtr uint32, valLen uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(keyPtr, keyLen)
		if !ok {
			return
		}
		key := string(bytes)

		bytes, ok = mem.Read(valPtr, valLen)
		if !ok {
			return
		}
		val := string(bytes)

		fmt.Printf("SET request %s for %s\n", key, val)

		instanceId := ctx.Value("instanceId").(string)

		handlerMap.CallHandler(wasmevents.LOG, instanceId, key, val)
	}
}
