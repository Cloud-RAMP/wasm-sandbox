package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Cloud-RAMP/wasm-sandbox/internal/modulestore"
	"github.com/Cloud-RAMP/wasm-sandbox/pkg/handlers"
)

// make the main sandbox functions that we expose here
func main() {
	ctx := context.Background()

	wasmBytes, err := os.ReadFile("./example/build/release.wasm")
	if err != nil {
		fmt.Println("Failed to read wasm file", err)
		return
	}

	store, err := modulestore.NewSandboxStore()
	if err != nil {
		fmt.Println("Failed to make sandbox store", err)
		return
	}

	if err := store.LoadModule("first", wasmBytes); err != nil {
		fmt.Println("Failed to load module", err)
		return
	}

	store.ExecuteOnModule(ctx, "first", handlers.ON_MESSAGE, "hello, world!")
}
