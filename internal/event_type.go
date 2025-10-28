package internal

type EventType int64

const (
	Join EventType = iota
	Leave
	Message
	Latency
)
