package loader

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero"
)

type LoaderFunction func(context.Context, string) ([]byte, error)

var loader LoaderFunction = nil

// Hook to be called on project initialization by the developer who accesses this
func SetLoaderFunction(function LoaderFunction) {
	loader = function
}

func LoadCompiled(ctx context.Context, runtime wazero.Runtime, moduleId string) (wazero.CompiledModule, error) {
	if loader == nil {
		return nil, fmt.Errorf("Loader function is not defined!")
	}

	bytes, err := loader(ctx, moduleId)
	if err != nil {
		return nil, err
	}

	compiled, err := runtime.CompileModule(ctx, bytes)
	if err != nil {
		return nil, err
	}

	return compiled, nil
}
