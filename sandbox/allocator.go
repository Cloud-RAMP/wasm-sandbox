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
func createASString(module api.Module, str string) (uint32, error) {
	ctx := context.Background()

	// Check for required runtime functions
	__new := module.ExportedFunction("__new")
	if __new == nil {
		return 0, fmt.Errorf("__new not exported (compile with --exportRuntime)")
	}

	// Convert to UTF-16 LE
	utf16Data := encodeUTF16LE(str)

	// AssemblyScript string memory layout:
	// - 4 bytes: type ID (rtid, typically 2 for strings)
	// - 4 bytes: length (in characters)
	// - N*2 bytes: UTF-16 data
	totalSize := uint64(8 + len(utf16Data))

	// Allocate memory with __new(size, id)
	results, err := __new.Call(ctx, totalSize, uint64(STRING_TYPE_ID))
	if err != nil {
		return 0, fmt.Errorf("__new failed: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("__new returned no result")
	}

	ptr := uint32(results[0])
	memory := module.Memory()

	// Write type ID (already set by __new? Sometimes you need to write it)
	typeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(typeBytes, STRING_TYPE_ID)
	memory.Write(ptr, typeBytes)

	// Write length (in characters)
	lenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBytes, uint32(len(str)))
	memory.Write(ptr, lenBytes)

	// Write UTF-16 data
	if !memory.Write(ptr+4, utf16Data) {
		return 0, fmt.Errorf("failed to write string data")
	}

	fmt.Printf("Creating string: %s\n", str)
	fmt.Printf("UTF-16 data: %v\n", utf16Data)
	fmt.Printf("Allocated pointer: %d, Total size: %d\n", ptr, totalSize)

	return ptr, nil
}
