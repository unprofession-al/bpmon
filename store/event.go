package store

import (
	"time"

	"github.com/unprofession-al/bpmon/status"
)

// Event represents a status change.
type Event struct {
	ID         ID                `json:"id"`
	Status     status.Status     `json:"status"`
	Annotation string            `json:"annotation"`
	Time       time.Time         `json:"time"`
	Tags       map[string]string `json:"tags"`
}

// SetID adds an ID to a `Span` based on its Time and Tags which sould be unique.
func (e *Event) SetID() {
	e.ID = NewID(e.Time, e.Tags)
}
