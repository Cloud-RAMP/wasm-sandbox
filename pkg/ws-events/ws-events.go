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

func (e WSEventType) String() string {
	return exportedWSEvents[e]
}
