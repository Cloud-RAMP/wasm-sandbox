package asmscript

import (
	"encoding/binary"
	"fmt"

	wsevents "github.com/Cloud-RAMP/wasm-sandbox/pkg/ws-events"
)

/*
Our protocol for communicating with WASM is:
* Fist we will have 4 bytes representing the number of strings passed in
* Each string item will be laid out as the following
  - 2 bytes for a +/- describing success / failure
  - asmscript strings are 2 byte aligned
  - 4 bytes for an integer describing the string length
  - The string
*/
func encodeArray(arr []string) []byte {
	count := uint32(len(arr))
	buf := make([]byte, 0, 5+count*8) // initial cap, will grow as needed

	// success indicator
	buf = append(buf, '+', 0)

	// number of strings
	buf = binary.LittleEndian.AppendUint32(buf, count)

	for _, s := range arr {
		buf = binary.LittleEndian.AppendUint32(buf, uint32(len(s)))
		buf = append(buf, s...)
	}

	return buf
}

// A WS event will just be encoded as an array of fields,
// it will be assumed that they are in the same order every time
func encodeWSEvent(event wsevents.WSEventInfo) []byte {
	feilds := []string{
		event.ConnectionId,
		event.RoomId,
		fmt.Sprint(event.Timestamp),
		event.Payload,
	}

	return encodeArray(feilds)
}

func WriteWSEvent(mod *ModuleContext, event wsevents.WSEventInfo) (uint64, uint64, error) {
	bytes := encodeWSEvent(event)
	return writeHelper(mod, bytes)
}

func WriteArray(mod *ModuleContext, array []string) (uint64, uint64, error) {
	bytes := encodeArray(array)
	return writeHelper(mod, bytes)
}
