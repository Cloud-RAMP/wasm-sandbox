package builder

import (
	"context"
	"log"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/asmscript"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/events"
	"github.com/tetratelabs/wazero/api"
)

func getAbortHandler(_ chan events.Event) any {
	return func(ctx context.Context, mod api.Module, messagePtr uint32, fileNamePtr uint32, line uint32, column uint32) {
		if mod != nil {
			message := asmscript.ReadASString(mod.Memory(), messagePtr)
			fileName := asmscript.ReadASString(mod.Memory(), fileNamePtr)
			log.Printf("AssemblyScript abort: %s at %s:%d:%d", message, fileName, line, column)
		}
	}
}

// Return a closure so that we can still send events to the eventChan
func getBroadcastHandler(eventChan chan events.Event) any {
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

		// set to events.BROADCAST right now, but we will need to support many different types of events coming from the SDK
		// One idea is to set 4 bytes of the response to an event type, and the rest to the event payload
		// That way we can differentiate events on this end
		eventChan <- events.Event{
			Type:       events.BROADCAST,
			Payload:    message,
			InstanceId: instanceId,
		}
	}
}
