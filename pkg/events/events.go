package events

import (
	"fmt"
)

type EventType int

const (
	ABORT EventType = iota
	BROADCAST
	SET
	GET
	LOG
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
	"log",
}

func (e EventType) String() string {
	return eventStrings[e]
}

type HandlerFunction func(EventType, string, ...string) (string, error)
type HandlerMap map[EventType]HandlerFunction

func NewHandlerMap() *HandlerMap {
	m := make(HandlerMap)
	return &m
}

// Returns itself so that it can be chained
func (m *HandlerMap) AddHandler(event EventType, handler HandlerFunction) *HandlerMap {
	(*m)[event] = handler
	return m
}

func (m *HandlerMap) CallHandler(event EventType, instanceId string, args ...string) (string, error) {
	h, ok := (*m)[event]
	if !ok {
		return "", fmt.Errorf("No handler present for %s event", event.String())
	}

	return h(event, instanceId, args...)
}
