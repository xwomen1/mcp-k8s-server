// internal/domain/event.go
package domain

import "time"

// (Normal or Warning)
type EventType string

const (
	EventTypeNormal  EventType = "Normal"
	EventTypeWarning EventType = "Warning"
)

type EventName string

type Event struct {
	Name            EventName `json:"name"`
	Namespace       Namespace `json:"namespace"`
	Type            EventType `json:"type"` // Normal, Warning
	Reason          string    `json:"reason"`
	Message         string    `json:"message"`
	SourceComponent string    `json:"source_component"`
	SourceHost      string    `json:"source_host"`
	InvolvedKind    string    `json:"involved_kind"` // (Pod, Deployment, Node...)
	InvolvedName    string    `json:"involved_name"`
	InvolvedUID     string    `json:"involved_uid"`
	Count           int32     `json:"count"`
	FirstTimestamp  time.Time `json:"first_timestamp"`
	LastTimestamp   time.Time `json:"last_timestamp"`
}
