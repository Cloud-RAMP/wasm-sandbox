package wasmevents

import (
	"fmt"
)

type WASMEventType int

const (
	ABORT WASMEventType = iota
	BROADCAST
	SET
	GET
	LOG
	DEBUG
	GET_USERS
)

type WASMEventInfo struct {
	// The unique ID of the connection sending this message
	ConnectionId string `json:"connection_id"`

	// The room to which the user is sending the message.
	// Maybe this can be optional if the application doesn't use rooms?
	RoomId string `json:"room_id"`

	// The unique ID of the application that this event is being sent to
	InstanceId string `json:"instance_id"`

	EventType WASMEventType `jsonL:"event_type"`
	Payload   []string      `json:"payload"`

	// The unix millisecond timestamp of the message
	Timestamp int64 `json:"timestamp"`
}

var eventStrings = [...]string{
	"abort",
	"broadcast",
	"set",
	"get",
	"log",
	"debug",
	"getUsers",
}

func (e WASMEventType) String() string {
	return eventStrings[e]
}

type HandlerFunction func(WASMEventInfo) (string, error)
type HandlerMap map[WASMEventType]HandlerFunction

func NewHandlerMap() *HandlerMap {
	m := make(HandlerMap)
	return &m
}

// Returns itself so that it can be chained
func (m *HandlerMap) AddHandler(event WASMEventType, handler HandlerFunction) *HandlerMap {
	(*m)[event] = handler
	return m
}

func (m *HandlerMap) CallHandler(event WASMEventInfo) (string, error) {
	h, ok := (*m)[event.EventType]
	if !ok {
		return "", fmt.Errorf("No handler present for %s event", event.EventType.String())
	}

	return h(event)
}
