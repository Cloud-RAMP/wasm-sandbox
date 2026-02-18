package events

type EventType int

const (
	ABORT EventType = iota
	BROADCAST
)

type Event struct {
	Payload    string
	Type       EventType
	InstanceId string
}

var eventStrings = [...]string{
	"abort",
	"broadcast",
}

func (e EventType) String() string {
	return eventStrings[e]
}
