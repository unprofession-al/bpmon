package store

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/status"
)

type Event struct {
	ID              EventID           `json:"id"`
	Status          status.Status     `json:"status"`
	Pseudo          bool              `json:"pseudo"`
	Annotation      string            `json:"annotation"`
	Start           time.Time         `json:"start"`
	End             time.Time         `json:"end"`
	Duration        float64           `json:"duration"`
	DurationPercent float64           `json:"duration_percent"`
	Tags            map[string]string `json:"tags"`
}

func (e *Event) SetEventID() {
	var pairs []string
	for key, value := range e.Tags {
		pairs = append(pairs, key+pairSeparator+value)
	}
	s := fmt.Sprintf("%v%s%s", e.Start.UnixNano(), timeTagSeparator, strings.Join(pairs, tagSeparator))
	e.ID = EventID(base64.RawURLEncoding.EncodeToString([]byte(s)))
}

type EventID string

const (
	pairSeparator    = "="
	tagSeparator     = ";"
	timeTagSeparator = " "
)

func (eid EventID) GetResultSet() (ResultSet, error) {
	rs := ResultSet{
		Tags: make(map[string]string),
	}
	data, err := base64.RawURLEncoding.DecodeString(string(eid))
	if err != nil {
		return rs, err
	}

	elements := strings.SplitN(string(data), timeTagSeparator, 2)
	if len(elements) != 2 {
		return rs, errors.New("Malformed Event ID")
	}

	nanos, err := strconv.ParseInt(elements[0], 10, 64)
	if err != nil {
		return rs, err
	}

	rs.Start = time.Unix(0, nanos)

	tags := strings.Split(elements[1], tagSeparator)
	for _, pair := range tags {
		touple := strings.SplitN(pair, pairSeparator, 2)
		if len(touple) != 2 {
			return rs, errors.New("Malformed Event ID")
		}
		rs.Tags[touple[0]] = touple[1]
	}

	return rs, nil
}
