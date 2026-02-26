package loader

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// What we need to do here is expose a "Load module" function, where we basically just give it an instanceID and it fetches it from some external data store.
// To make the systems less tightly coupled, we could define a handler function the user can specify

type LoaderFunction func(string) ([]byte, error)

var loader LoaderFunction = nil

func SetLoaderFunction(function LoaderFunction) {
	loader = function
}

func Load(ctx context.Context, runtime wazero.Runtime, moduleId string) (api.Module, error) {
	if loader == nil {
		return nil, fmt.Errorf("Loader function is not defined!")
	}

	bytes, err := loader(moduleId)
	if err != nil {
		return nil, err
	}

	module, err := runtime.Instantiate(ctx, bytes)
	if err != nil {
		return nil, err
	}

	return module, nil
}
