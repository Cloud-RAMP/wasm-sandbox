package asmscript

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero/api"
)

type ModuleContext struct {
	Module api.Module
	Ctx    context.Context
	Mu     *sync.Mutex
}

// Encode string to UTF-16 Little Endian bytes
func encodeUTF16LE(s string) []byte {
	bytes := make([]byte, len(s)*2)
	for i, r := range s {
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(r))
	}
	return bytes
}

// Decode UTF-16 Little Endian bytes to string
func decodeUTF16LE(b []byte) string {
	runes := make([]rune, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		if i+1 < len(b) {
			runes[i/2] = rune(binary.LittleEndian.Uint16(b[i:]))
		}
	}
	return string(runes)
}

// Write a string to module memory
func writeHelper(mod *ModuleContext, bytes []byte) (uint64, uint64, error) {
	__new := mod.Module.ExportedFunction("__new")
	if __new == nil {
		return 0, 0, fmt.Errorf("__new not exported")
	}

	results, err := __new.Call(mod.Ctx, uint64(len(bytes)), 0)
	if err != nil {
		return 0, 0, fmt.Errorf("__new failed: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("__new returned no result")
	}

	// __new returns the pointer value
	ptr := uint32(results[0])
	memory := mod.Module.Memory()

	// Write UTF-16 data
	mod.Mu.Lock()
	if !memory.Write(ptr, bytes) {
		mod.Mu.Unlock()
		return 0, 0, fmt.Errorf("failed to write string data")
	}
	mod.Mu.Unlock()

	return uint64(ptr), uint64(len(bytes)), nil
}
