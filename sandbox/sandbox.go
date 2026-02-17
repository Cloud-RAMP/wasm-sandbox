package sandbox

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type Event struct {
	Type    string
	Payload string
}

var invocation struct {
	Events []Event
}

func callOnMessage(module api.Module, message string) {
	// Write the string to memory
	ptr, err := createASString(module, message)
	if err != nil {
		log.Fatalf("Failed to create WASM string %v", err)
		return
	}

	// Call the `onMessage` function with the pointer and length
	onMessage := module.ExportedFunction("__onMessage")
	_, err = onMessage.Call(context.Background(), uint64(ptr), uint64(len(message)))
	if err != nil {
		log.Fatalf("Failed to call onMessage: %v", err)
	}
}

// Actually execute user code and enforce limits here
func ExecuteSandbox(ctx context.Context) {
	wasmBytes, err := os.ReadFile("./example/build/release.wasm")
	if err != nil {
		fmt.Println("Failed to read wasm file", err)
		return
	}

	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithMemoryLimitPages(10))
	defer runtime.Close(ctx)

	hostModuleBuilder := runtime.NewHostModuleBuilder("env")

	// amscript requires abort function to be present
	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, messagePtr uint32, fileNamePtr uint32, line uint32, column uint32) {
			if mod != nil {
				message := readASString(mod.Memory(), messagePtr)
				fileName := readASString(mod.Memory(), fileNamePtr)
				log.Printf("AssemblyScript abort: %s at %s:%d:%d", message, fileName, line, column)
			} else {
				log.Printf("AssemblyScript abort called")
			}
		}).
		Export("abort")

	hostModuleBuilder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, ptr uint32, len uint32) {
			// Get memory from the module parameter
			mem := mod.Memory()
			if mem == nil {
				log.Printf("No memory available")
				return
			}

			// Read the string from memory
			bytes, ok := mem.Read(ptr, len)
			if !ok {
				log.Printf("Failed to read memory at ptr=%d, len=%d", ptr, len)
				return
			}
			message := string(bytes)

			// Capture event
			invocation.Events = append(invocation.Events, Event{
				Type:    "broadcast",
				Payload: message,
			})

		}).
		Export("broadcast")

	_, err = hostModuleBuilder.Instantiate(ctx)
	if err != nil {
		log.Fatalf("Failed to instantiate host module: %v", err)
	}

	module, err := runtime.Instantiate(ctx, wasmBytes)
	if err != nil {
		log.Fatalf("Failed to instantiate WASM module: %v", err)
	}
	defer module.Close(ctx)

	for f := range module.ExportedFunctionDefinitions() {
		fmt.Printf("Exported function: %s\n", f)
	}
	callOnMessage(module, "hello world")

	fmt.Println("WebAssembly complete")
	fmt.Println("Invocations:", invocation.Events)
}

func ExecuteSandboxWithProtection() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use defer/recover to catch panics
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Sandbox panicked: %v\n", r)
		}
	}()

	// Reset events
	invocation.Events = nil

	// Call the main execution
	ExecuteSandbox(ctx)
}
