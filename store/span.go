package store

import (
	"time"

	"github.com/unprofession-al/bpmon/status"
)

// Span represents the time span between two events of different status.
type Span struct {
	ID              ID              `json:"id"`
	Status          status.Status   `json:"status"`
	Pseudo          bool            `json:"pseudo"`
	Annotation      string          `json:"annotation"`
	Start           time.Time       `json:"start"`
	End             time.Time       `json:"end"`
	Duration        float64         `json:"duration"`
	DurationPercent float64         `json:"duration_percent"`
	Tags            map[Kind]string `json:"tags"`
}

// SetID adds an ID to a `Span` based on its Start and Tags which sould be unique.
func (s *Span) SetID() {
	s.ID = NewID(s.Start, s.Tags)
}
