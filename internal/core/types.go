// Package core contains the standard event types used by thingsboard-gateway
package core

type EventType string

const (
	EventConnect    EventType = "connect"
	EventDisconnect EventType = "disconnect"
	EventAttribute  EventType = "attributes"
	EventRPC        EventType = "rpc"
	EventUnknown    EventType = "unknown"
)
