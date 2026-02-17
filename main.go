package main

import (
	"github.com/Cloud-RAMP/wasm-sandbox/sandbox"
)

// make the main sandbox functions that we expose here
func main() {
	sandbox.ExecuteSandboxWithProtection()
}
