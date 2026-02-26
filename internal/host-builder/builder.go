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

	// LOG
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(logHandler(handlerMap)).
		Export(wasmevents.LOG.String())

	// DEBUG
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(debugHandler(handlerMap)).
		Export(wasmevents.DEBUG.String())

	// GET_USERS
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(getUsersHandler(handlerMap)).
		Export(wasmevents.GET_USERS.String())

	// SEND_MESSAGE
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(sendMessageHandler(handlerMap)).
		Export(wasmevents.SEND_MESSAGE.String())

	// FETCH
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(fetchHandler(handlerMap)).
		Export(wasmevents.FETCH.String())

	return hostModuleBuilder.Instantiate(context.Background())
}
