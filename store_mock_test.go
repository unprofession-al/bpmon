package bpmon

import (
	"time"

	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

type StoreMock struct{}

func (pp StoreMock) Write(p *store.ResultSet) error {
	return nil
}

func (pp StoreMock) GetLatest(rs store.ResultSet) (store.ResultSet, error) {
	return store.ResultSet{}, nil
}

func (pp StoreMock) GetSpans(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Span, error) {
	return []store.Span{}, nil
}

func (pp StoreMock) AnnotateEvent(id store.ID, annotation string) (store.ResultSet, error) {
	return store.ResultSet{}, nil
}

func (pp StoreMock) GetEvents(kind store.Kind, start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Event, error) {
	return []store.Event{}, nil
}
