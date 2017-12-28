package store

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/status"
)

type Event struct {
	ID              string            `json:"id"`
	Status          status.Status     `json:"status"`
	Pseudo          bool              `json:"pseudo"`
	Annotation      string            `json:"annotation"`
	Start           time.Time         `json:"start"`
	End             time.Time         `json:"end"`
	Duration        float64           `json:"duration"`
	DurationPercent float64           `json:"duration_percent"`
	Tags            map[string]string `json:"tags"`
}

const (
	pairSeparator    = "="
	tagSeparator     = ";"
	timeTagSeparator = " "
)

func (e *Event) SetID() {
	var pairs []string
	for key, value := range e.Tags {
		pairs = append(pairs, key+pairSeparator+value)
	}
	s := fmt.Sprintf("%v%s%s", e.Start.UnixNano(), timeTagSeparator, strings.Join(pairs, tagSeparator))
	e.ID = base64.RawURLEncoding.EncodeToString([]byte(s))
}
