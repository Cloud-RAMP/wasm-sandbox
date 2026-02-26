package loader

import (
	"fmt"
	"os"
)

func init() {
	SetLoaderFunction(MockLoaderFunction)
}

// To be used in testing
func MockLoaderFunction(moduleId string) ([]byte, error) {
	wasmBytes, err := os.ReadFile(moduleId)
	if err != nil {
		fmt.Println("Failed to read wasm file", err)
		return nil, err
	}

	return wasmBytes, nil
}
