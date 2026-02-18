package events

type EventType int

const (
	ABORT EventType = iota
	BROADCAST
	SET
	GET
)

type Event struct {
	Payload    string
	Type       EventType
	InstanceId string
}

var eventStrings = [...]string{
	"abort",
	"broadcast",
	"set",
	"get",
}

func (e EventType) String() string {
	return eventStrings[e]
}
