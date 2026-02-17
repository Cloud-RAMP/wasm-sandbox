package sandbox

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

const STRING_TYPE_ID = 2

// Encode string to UTF-16 LE bytes
func encodeUTF16LE(s string) []byte {
	bytes := make([]byte, len(s)*2)
	for i, r := range s {
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(r))
	}
	return bytes
}

// Decode UTF-16 LE bytes to string
func decodeUTF16LE(b []byte) string {
	runes := make([]rune, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		if i+1 < len(b) {
			runes[i/2] = rune(binary.LittleEndian.Uint16(b[i:]))
		}
	}
	return string(runes)
}

// Read AssemblyScript string from memory
func readASString(mem api.Memory, ptr uint32) string {
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

// Create an AssemblyScript string in the module's memory
func createASString(module api.Module, str string) (uint64, uint64, error) {
	ctx := context.Background()

	// Check that the runtime function exists
	__new := module.ExportedFunction("__new")
	if __new == nil {
		return 0, 0, fmt.Errorf("__new not exported")
	}

	// Convert to UTF-16 Little Endian
	utf16Data := encodeUTF16LE(str)
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
