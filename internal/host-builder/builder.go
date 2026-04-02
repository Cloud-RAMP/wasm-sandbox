package hostbuilder

import (
	"context"

	wasmevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/wasm-events"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func BuildHostModule(ctx context.Context, runtime wazero.Runtime, handlerMap *wasmevents.HandlerMap) (api.Module, error) {
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
		WithFunc(setHandler(handlerMap, wasmevents.SET)).
		Export(wasmevents.SET.String())

	// DB_SET function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(setHandler(handlerMap, wasmevents.DB_SET)).
		Export(wasmevents.DB_SET.String())

	// GET function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(getHandler(handlerMap, wasmevents.GET)).
		Export(wasmevents.GET.String())

	// DB_GET function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(getHandler(handlerMap, wasmevents.DB_GET)).
		Export(wasmevents.DB_GET.String())

	// DEL function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(delHandler(handlerMap)).
		Export(wasmevents.DEL.String())

	// DB_DEL function
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(dbDelHandler(handlerMap)).
		Export(wasmevents.DB_DEL.String())

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

	// CLOSE_CONNECTION
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(closeConnectionHanlder(handlerMap)).
		Export(wasmevents.CLOSE_CONNECTION.String())

	// FETCH
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(fetchHandler(handlerMap)).
		Export(wasmevents.FETCH.String())

	return hostModuleBuilder.Instantiate(ctx)
}
