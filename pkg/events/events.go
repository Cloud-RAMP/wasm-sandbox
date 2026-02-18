package events

type EventType int

const (
	ABORT EventType = iota
	BROADCAST
)

type Event struct {
	Payload string
	Type    EventType
}
