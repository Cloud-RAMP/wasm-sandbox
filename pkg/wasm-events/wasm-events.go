package wasmevents

import (
	"fmt"

	modulelocks "github.com/Cloud-RAMP/wasm-sandbox/internal/module-locks"
)

type WASMEventType int

const (
	// Default "ABORT" method called by WASM when something fails
	ABORT WASMEventType = iota

	// Broadcast a message from a user to a room
	BROADCAST

	// Set a key in an in-memory KV store
	//
	// Do we distribute this across all instances, or just within the machine?
	SET

	// Get a value from in-memory KV store
	//
	// Same question as above
	GET

	// Delete a key in the in-memory KV store
	DEL

	// Set a key/value in the persistent storage
	DB_SET

	// Get a key/value in the persistent storage
	DB_GET

	// Delete key in persistent storage
	DB_DEL

	// Log something (in the user application)
	LOG

	// Debug, for development purposes
	DEBUG

	// Get a list of all users in the current room
	GET_USERS

	// Send a message to a specific user
	SEND_MESSAGE

	// Send an HTTP request to a given URL with a request type and body
	FETCH
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
	"del",
	"dbSet",
	"dbGet",
	"dbDel",
	"log",
	"debug",
	"getUsers",
	"sendMessage",
	"fetch",
}

func (e WASMEventType) String() string {
	return eventStrings[e]
}

type HandlerFunction func(*WASMEventInfo) (string, error)
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

func (m *HandlerMap) CallHandler(event *WASMEventInfo) (string, error) {
	h, ok := (*m)[event.EventType]
	if !ok {
		return "", fmt.Errorf("No handler present for %s event", event.EventType.String())
	}

	// give up lock control while some external event is called
	// (this may cause concurrency issues?)
	modulelocks.Unlock(event.InstanceId)
	res, err := h(event)
	modulelocks.Lock(event.InstanceId)

	return res, err
}
