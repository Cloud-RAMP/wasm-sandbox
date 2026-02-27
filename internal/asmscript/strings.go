package asmscript

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

// string types in asmscript are defined as 2
const STRING_TYPE_ID = 2

// Read AssemblyScript string from memory
func ReadASString(mem api.Memory, ptr uint32) string {
	// Read length prefix (bytes 4-7)
	lenBytes, ok := mem.Read(ptr+4, 4)
	if !ok {
		return "<failed to read string>"
	}
	strLen := binary.LittleEndian.Uint32(lenBytes)

	// Read UTF-16 data (starts at offset 8)
	data, ok := mem.Read(ptr+8, strLen*2)
	if !ok {
		return "<failed to read string data>"
	}

	return decodeUTF16LE(data)
}

// A message string with + denotes a successful string
func CreateASString(module api.Module, str string) (uint64, uint64, error) {
	return createStringInternal(module, str, '+')
}

// A message starting with - denotes an error
func CreateASError(module api.Module, err error) (uint64, uint64, error) {
	return createStringInternal(module, err.Error(), '-')
}

// Create a string in the given module's memory
//
// Return string location, string length, and possible error
func createStringInternal(module api.Module, str string, indicator rune) (uint64, uint64, error) {
	ctx := context.Background()

	// Check that the runtime function exists
	__new := module.ExportedFunction("__new")
	if __new == nil {
		return 0, 0, fmt.Errorf("__new not exported")
	}

	// Convert to UTF-16 Little Endian
	utf16Data := encodeUTF16LE(str)
	utf16Data = append([]byte{byte(indicator), 0}, utf16Data...)
	totalSize := uint64(len(utf16Data))

	// Allocate memory with __new(size, id)
	// __new creates the headers for data type and data size
	results, err := __new.Call(ctx, totalSize, uint64(STRING_TYPE_ID))
	if err != nil {
		return 0, 0, fmt.Errorf("__new failed: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("__new returned no result")
	}

	// __new returns the pointer value
	ptr := uint32(results[0])
	memory := module.Memory()

	// Write UTF-16 data
	if !memory.Write(ptr, utf16Data) {
		return 0, 0, fmt.Errorf("failed to write string data")
	}

	return uint64(ptr), totalSize, nil
}
