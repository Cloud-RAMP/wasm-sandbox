# WASM-Sandbox

This repo contains the code necessary for everything related to user code execution in our cloud-RAMP project.

### Module Goals

The goals of this module are as follows:
* Load user-uploaded code that has been compiled to WASM from some store that will hold this data
    * Currently, this only supports AssemblyScript (a dialect of typescript)
        * We will need to create an sdk for programmers, likley in a separate repo
    * Likely a cloud-bucket store provider. Code should be accessible by some unique identifier
* Keep this WASM module stored in a fixed size table in memory (cache on each server)
    * Old modules that have not been requested recently should be evicted with an LRU policy
        * This should be doable, as each module will have an associated LastRequest time field
* Support function calls like "onMessage", "onJoin", "onLeave" from external service


### APIs

The main APIs that we should expose to the main program will consist mostly of these functions, assuming that if the module has not been loaded, it will be done by the internals of this project.
These functions will also emit events, which we will need to define a consistent syntax for. This will likely be a list of structs as follows:
```go
type Event struct {
    Payload   string      // actual data being sent from the user's function
    EventType EventType   // some enum describing the different events we support
    Service   string      // the service that created this event. this might not be completely necessary, since the code that calls this should know what service it belongs to.
}
```
There are two possible ways to emit these events:
* In a readonly chan that is returned from the called function
* A list that is returned from the called function

I think the chan option is the better suited one here, since events may be returned asynchronously, as opposed to all at once