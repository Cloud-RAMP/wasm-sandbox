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
)

type WASMEvent struct {
	Payload    string
	Type       WASMEventType
	InstanceId string
}

var eventStrings = [...]string{
	"abort",
	"broadcast",
	"set",
	"get",
	"log",
}

func (e WASMEventType) String() string {
	return eventStrings[e]
}

type HandlerFunction func(WASMEventType, string, ...string) (string, error)
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

func (m *HandlerMap) CallHandler(event WASMEventType, instanceId string, args ...string) (string, error) {
	h, ok := (*m)[event]
	if !ok {
		return "", fmt.Errorf("No handler present for %s event", event.String())
	}

	fmt.Printf("Event: %s, instance: %s, args: %v\n", event.String(), instanceId, args)

	return h(event, instanceId, args...)
}
