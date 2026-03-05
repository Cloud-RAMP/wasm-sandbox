package hostbuilder

import (
	"context"
	"fmt"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	modulelocks "github.com/Cloud-RAMP/wasm-sandbox/internal/module-locks"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

// These functions all return closures that capture the value of events.HandlerMap becuause it avoids circular import issues

// Returns a boilerplate WASMEvent object injected with variables from the CTX
func getWASMEvent(ctx context.Context, eventType wasmevents.WASMEventType, payload ...string) wasmevents.WASMEventInfo {
	instanceId, ok := ctx.Value("instanceId").(string)
	if !ok {
		logging.Logger.Errorf("Failed to parse instanceID from ctx in getWASMEvent")
		return wasmevents.WASMEventInfo{} // fix this, issue #14
	}

	connectionId, ok := ctx.Value("connectionId").(string)
	if !ok {
		logging.Logger.Errorf("Failed to parse connectionId from ctx in getWASMEvent")
		return wasmevents.WASMEventInfo{}
	}

	roomId, ok := ctx.Value("roomId").(string)
	if !ok {
		logging.Logger.Errorf("Failed to parse roomId from ctx in getWASMEvent")
		return wasmevents.WASMEventInfo{}
	}

	return wasmevents.WASMEventInfo{
		ConnectionId: connectionId,
		InstanceId:   instanceId,
		RoomId:       roomId,
		Timestamp:    time.Now().UnixMilli(),
		EventType:    eventType,
		Payload:      payload,
	}
}

func getModuleContext(ctx context.Context, mod api.Module) *asmscript.ModuleContext {
	instanceId, ok := ctx.Value("instanceId").(string)
	if !ok {
		logging.Logger.Errorf("Failed to parse instanceID from ctx in getModuleContext")
		return nil
	}

	return &asmscript.ModuleContext{
		Ctx:    ctx,
		Module: mod,
		Mu:     modulelocks.GetLockReference(instanceId),
	}
}

func abortHandler(_ *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, messagePtr uint32, fileNamePtr uint32, line uint32, column uint32) {
		if mod != nil {
			message := asmscript.ReadASString(mod.Memory(), messagePtr)
			fileName := asmscript.ReadASString(mod.Memory(), fileNamePtr)
			logging.Logger.Errorf("AssemblyScript abort: %s at %s:%d:%d", message, fileName, line, column)
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

		event := getWASMEvent(ctx, wasmevents.BROADCAST, string(bytes))
		_, _ = handlerMap.CallHandler(event)
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

		event := getWASMEvent(ctx, wasmevents.SET, key, val)
		_, err := handlerMap.CallHandler(event)
		if err != nil {
			//some error handling here
		}
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
		event := getWASMEvent(ctx, wasmevents.GET, string(bytes))

		// do some sort of redis operation here, up to the external handler
		val, err := handlerMap.CallHandler(event)
		if err != nil {
			ptr, _, _ := asmscript.CreateASError(
				getModuleContext(ctx, mod),
				fmt.Errorf("Failed to execute GET hander"),
			)
			return uint32(ptr)
		}

		// insert the string into module memory
		ptr, _, err := asmscript.CreateASString(
			getModuleContext(ctx, mod),
			val,
		)
		if err != nil {
			return 0
		}

		return uint32(ptr)
	}
}

func logHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, strPtr uint32, strLen uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(strPtr, strLen)
		if !ok {
			return
		}
		info := string(bytes)

		event := getWASMEvent(ctx, wasmevents.LOG, info)
		handlerMap.CallHandler(event)
	}
}

func debugHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, strPtr uint32, strLen uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(strPtr, strLen)
		if !ok {
			return
		}
		info := string(bytes)

		event := getWASMEvent(ctx, wasmevents.DEBUG, info)
		handlerMap.CallHandler(event)
	}
}

func getUsersHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module) uint32 {
		event := getWASMEvent(ctx, wasmevents.GET_USERS, "")
		handlerMap.CallHandler(event)

		ptr, _, err := asmscript.CreateASError(
			getModuleContext(ctx, mod),
			fmt.Errorf("testing error"),
		)
		if err != nil {
			ptr, _, _ := asmscript.CreateASError(
				getModuleContext(ctx, mod),
				fmt.Errorf("Failed to execute GET_USERS hander"),
			)
			return uint32(ptr)
		}

		return uint32(ptr)
	}
}

func sendMessageHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, userPtr uint32, userLen uint32, msgPtr uint32, msgLen uint32) {
		mem := mod.Memory()
		if mem == nil {
			return
		}

		bytes, ok := mem.Read(userPtr, userLen)
		if !ok {
			return
		}
		user := string(bytes)

		bytes, ok = mem.Read(msgPtr, msgLen)
		if !ok {
			return
		}
		msg := string(bytes)

		event := getWASMEvent(ctx, wasmevents.SEND_MESSAGE, user, msg)
		handlerMap.CallHandler(event)
	}
}

func fetchHandler(handlerMap *wasmevents.HandlerMap) any {
	return func(ctx context.Context, mod api.Module, urlPtr uint32, urlLen uint32, methodPtr uint32, methodLen uint32, bodyPtr uint32, bodyLen uint32) uint32 {
		mem := mod.Memory()
		if mem == nil {
			return 0
		}

		bytes, ok := mem.Read(urlPtr, urlLen)
		if !ok {
			return 0
		}
		url := string(bytes)

		bytes, ok = mem.Read(methodPtr, methodLen)
		if !ok {
			return 0
		}
		method := string(bytes)

		bytes, ok = mem.Read(bodyPtr, bodyLen)
		if !ok {
			return 0
		}
		body := string(bytes)

		event := getWASMEvent(ctx, wasmevents.FETCH, url, method, body)
		_, err := handlerMap.CallHandler(event)
		if err != nil {
			ptr, _, _ := asmscript.CreateASError(
				getModuleContext(ctx, mod),
				fmt.Errorf("Failed to execute FETCH hander"),
			)
			return uint32(ptr)
		}

		// TODO: remove
		resp := "bruh"
		ptr, _, _ := asmscript.CreateASString(
			getModuleContext(ctx, mod),
			resp,
		)
		return uint32(ptr)
	}
}
