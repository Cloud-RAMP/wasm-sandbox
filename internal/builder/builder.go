package builder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/pkg/events"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func BuildHostModule(runtime wazero.Runtime, eventChan chan events.Event) (api.Module, error) {
	hostModuleBuilder := runtime.NewHostModuleBuilder("env")

	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(getAbortHandler(eventChan)).
		Export(events.ABORT.String())

	// Broadcast function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(getBroadcastHandler(eventChan)).
		Export(events.BROADCAST.String())

	return hostModuleBuilder.Instantiate(context.Background())
}
