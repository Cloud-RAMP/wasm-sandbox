package builder

import (
	"context"

	"github.com/Cloud-RAMP/wasm-sandbox/pkg/events"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func BuildHostModule(runtime wazero.Runtime, handlerMap *events.HandlerMap) (api.Module, error) {
	hostModuleBuilder := runtime.NewHostModuleBuilder("env")

	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(abortHandler(handlerMap)).
		Export(events.ABORT.String())

	// Broadcast function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(broadcastHandler(handlerMap)).
		Export(events.BROADCAST.String())

	// SET function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(setHandler(handlerMap)).
		Export(events.SET.String())

	// GET function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(getHandler(handlerMap)).
		Export(events.GET.String())

	return hostModuleBuilder.Instantiate(context.Background())
}
