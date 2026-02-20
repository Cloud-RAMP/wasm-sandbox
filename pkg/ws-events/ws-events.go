package wsevents

type WSEventType int

const (
	ON_MESSAGE WSEventType = iota
	ON_JOIN
	ON_LEAVE
)

// These are the function names that will be defined within our AssemblyScript SDK
var exportedWSEvents = [...]string{
	"__onMessage",
	"__onJoin",
	"__onLeave",
}

// This event defines info that will be sent INTO the WASM sandbox for the user to use in their code
type WSEventInfo struct {
	// The unique ID of the connection sending this message
	ConnectionId string `json:"connection_id"`

	// The room to which the user is sending the message.
	// Maybe this can be optional if the application doesn't use rooms?
	RoomId string `json:"room_id"`

	// The unique ID of the application that this event is being sent to
	InstanceId string `json:"instance_id"`

	// The type of event (message, join, leave, etc.) See WSEventType
	EventType WSEventType // not necessary to make JSON

	// The data sent with the event. This does not need to be populated
	Payload string `json:"payload"`

	// The unix millisecond timestamp of the message
	Timestamp int64 `json:"timestamp"`
}

func (e WSEventType) String() string {
	return exportedWSEvents[e]
}
