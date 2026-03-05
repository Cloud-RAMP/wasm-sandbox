package asmscript

import (
	"encoding/binary"

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
func CreateASString(mod *ModuleContext, str string) (uint64, uint64, error) {
	return createStringInternal(mod, str, '+')
}

// A message starting with - denotes an error
func CreateASError(mod *ModuleContext, err error) (uint64, uint64, error) {
	return createStringInternal(mod, err.Error(), '-')
}

// Create a string in the given module's memory
//
// Return string location, string length, and possible error
func createStringInternal(mod *ModuleContext, str string, indicator rune) (uint64, uint64, error) {

	// Convert to UTF-16 Little Endian
	utf16Data := encodeUTF16LE(str)
	utf16Data = append([]byte{byte(indicator), 0}, utf16Data...)

	return writeHelper(mod, utf16Data)
}
