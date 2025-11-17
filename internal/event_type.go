package internal

type EventType int64

const (
	Join EventType = iota + 1
	Leave
	Message
	Latency
)
