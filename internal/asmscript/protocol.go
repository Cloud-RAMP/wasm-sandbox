package asmscript

import (
	"context"
	"encoding/binary"
	"fmt"

	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
	"github.com/tetratelabs/wazero/api"
)

/*
Our protocol for communicating with WASM is:
* Fist we will have 4 bytes representing the number of strings passed in
* Each string item will be laid out as the following
  - 4 bytes for an integer describing the string length
  - The string
*/
func EncodeWSEvent(event wsevents.WSEventInfo) []byte {
	fields := []string{
		event.ConnectionId,
		event.RoomId,
		fmt.Sprint(event.Timestamp),
		event.Payload,
	}

	count := uint32(len(fields))
	buf := make([]byte, 0, 4+count*8) // initial cap, will grow as needed

	// Write the number of strings (4 bytes, little endian)
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, count)
	buf = append(buf, tmp...)

	for _, s := range fields {
		strBytes := []byte(s)
		binary.LittleEndian.PutUint32(tmp, uint32(len(strBytes)))
		buf = append(buf, tmp...)
		buf = append(buf, strBytes...)
	}

	return buf
}

func WriteWSEvent(module api.Module, event wsevents.WSEventInfo) (uint64, uint64, error) {
	bytes := EncodeWSEvent(event)
	ctx := context.Background()

	// Check that the runtime function exists
	__new := module.ExportedFunction("__new")
	if __new == nil {
		return 0, 0, fmt.Errorf("__new not exported")
	}

	results, err := __new.Call(ctx, uint64(len(bytes)), 0)
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
	if !memory.Write(ptr, bytes) {
		return 0, 0, fmt.Errorf("failed to write string data")
	}

	return uint64(ptr), uint64(len(bytes)), nil
}
