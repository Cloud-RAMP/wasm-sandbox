package loader

import (
	"context"
	"fmt"
	"os"
	"time"
)

func init() {
	SetLoaderFunction(MockLoaderFunction)
}

// To be used in testing
func MockLoaderFunction(ctx context.Context, moduleId string) ([]byte, error) {
	time.Sleep(1 * time.Second) // simulated delay (longer than probably normal)

	wasmBytes, err := os.ReadFile(moduleId)
	if err != nil {
		fmt.Println("Failed to read wasm file", err)
		return nil, err
	}

	return wasmBytes, nil
}
