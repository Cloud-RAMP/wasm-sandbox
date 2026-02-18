package asmscript

import "encoding/binary"

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
