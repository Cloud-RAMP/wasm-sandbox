package hostbuilder

import (
	"context"
	"fmt"
	"time"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	modulelocks "github.com/Cloud-RAMP/wasm-sandbox/internal/module-locks"
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

func getModuleContext(ctx context.Context, mod api.Module) (*asmscript.ModuleContext, error) {
	instanceId, ok := ctx.Value("instanceId").(string)
	if !ok {
		err := fmt.Errorf("Failed to parse instanceID from ctx in getModuleContext")
		return nil, err
	}

	lock := modulelocks.GetLockReference(instanceId)
	if lock == nil {
		err := fmt.Errorf("Failed to get lock for module %s", instanceId)
		return nil, err
	}

	return &asmscript.ModuleContext{
		Ctx:    ctx,
		Module: mod,
		Mu:     modulelocks.GetLockReference(instanceId),
	}, nil
}
