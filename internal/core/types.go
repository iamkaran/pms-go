package core

type EventType string

const (
	EventConnect    EventType = "connect"
	EventDisconnect EventType = "disconnect"
	EventAttributes EventType = "attributes"
	EventRPC        EventType = "rpc"
	EventUnknown    EventType = "unknown"
)
