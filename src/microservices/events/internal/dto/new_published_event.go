package dto

import "time"

type PublishedEvent struct {
	Details PublishedEventDetails
	Payload EventPayload
}

type EventPayload struct {
	ID        string
	Type      string
	Timestamp time.Time
	Payload   map[string]string
}

type PublishedEventDetails struct {
	Status    string
	Partition int64
	Offset    int64
}
