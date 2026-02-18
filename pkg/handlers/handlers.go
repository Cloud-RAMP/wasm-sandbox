package handlers

type Handler int

const (
	ON_MESSAGE Handler = iota
	ON_JOIN
	ON_LEAVE
)

// These are the function names that will be defined within our AssemblyScript SDK
var exportedHandlers = [...]string{
	"__onMessage",
	"__onJoin",
	"__onLeave",
}

func (e Handler) String() string {
	return exportedHandlers[e]
}
