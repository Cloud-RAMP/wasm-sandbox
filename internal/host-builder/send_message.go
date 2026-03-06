package hostbuilder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/logging"
	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"
	"github.com/tetratelabs/wazero/api"
)

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

		event, err := getWASMEvent(ctx, wasmevents.SEND_MESSAGE, user, msg)
		if event == nil {
			logging.Logger.Errorf("Failed to create WASM event in handler %s: %v", wasmevents.SEND_MESSAGE.String(), err)
			return
		}

		_, err = handlerMap.CallHandler(event)
		if err != nil {
			logging.Logger.Errorf("Failed to call handler in %s: %v", wasmevents.SEND_MESSAGE.String(), err)
		}
	}
}
