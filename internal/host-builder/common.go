package hostbuilder

import (
	"context"
	"fmt"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

// Returns a boilerplate WASMEvent object injected with variables from the CTX
func getWASMEvent(ctx context.Context, eventType wasmevents.WASMEventType, payload ...string) (*wasmevents.WASMEventInfo, error) {
	instanceId, ok := ctx.Value("instanceId").(string)
	if !ok {
		err := fmt.Errorf("Failed to parse instanceId from ctx in getWASMEvent")
		return nil, err
	}

	connectionId, ok := ctx.Value("connectionId").(string)
	if !ok {
		err := fmt.Errorf("Failed to parse connectionId from ctx in getWASMEvent")
		return nil, err
	}

	roomId, ok := ctx.Value("roomId").(string)
	if !ok {
		err := fmt.Errorf("Failed to parse roomId from ctx in getWASMEvent")
		return nil, err
	}

	return &wasmevents.WASMEventInfo{
		ConnectionId: connectionId,
		InstanceId:   instanceId,
		RoomId:       roomId,
		Timestamp:    time.Now().UnixMilli(),
		EventType:    eventType,
		Payload:      payload,
	}, nil
}

// Return a "ModuleContext" object to reduce boilerplate in handler code
func getModuleContext(ctx context.Context, mod api.Module) *asmscript.ModuleContext {
	return &asmscript.ModuleContext{
		Ctx:    ctx,
		Module: mod,
	}
}

type errorMessagesType int

const (
	MOD_MEMORY_ERR errorMessagesType = iota
	MEM_READ_ERR
	GET_WASM_EVENT_ERR
	CREATE_AS_STRING_ERR
	EXTERNAL_HANDLER_ERR
)

var errorMessages = [...]error{
	fmt.Errorf("Failed to access module memory"),
	fmt.Errorf("Failed to read module memory"),
	fmt.Errorf("Failed to parse event information"),
	fmt.Errorf("Failed to create string in WASM memory"),
	fmt.Errorf("Failed external call"),
}

func writeErrorMessage(mod *asmscript.ModuleContext, err errorMessagesType) uint32 {
	ptr, _, _ := asmscript.CreateASError(mod, errorMessages[err])
	return uint32(ptr)
}
