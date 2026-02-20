package hostbuilder

import (
	"context"

	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func BuildHostModule(runtime wazero.Runtime, handlerMap *wasmevents.HandlerMap) (api.Module, error) {
	hostModuleBuilder := runtime.NewHostModuleBuilder("env")

	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(abortHandler(handlerMap)).
		Export(wasmevents.ABORT.String())

	// Broadcast function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(broadcastHandler(handlerMap)).
		Export(wasmevents.BROADCAST.String())

	// SET function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(setHandler(handlerMap)).
		Export(wasmevents.SET.String())

	// GET function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(getHandler(handlerMap)).
		Export(wasmevents.GET.String())

	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(logHandler(handlerMap)).
		Export(wasmevents.LOG.String())

	return hostModuleBuilder.Instantiate(context.Background())
}
