package bpmon

import (
	"time"

	"github.com/unprofession-al/bpmon/store"
)

type StoreMock struct{}

func (pp StoreMock) Write(p *store.ResultSet) error {
	return nil
}

func (pp StoreMock) GetLatest(rs store.ResultSet) (store.ResultSet, error) {
	return store.ResultSet{}, nil
}

func (pp StoreMock) GetEvents(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration) ([]store.Event, error) {
	return []store.Event{}, nil
}
func (pp StoreMock) AnnotateEvent(id string, annotation string) (store.ResultSet, error) {
	return store.ResultSet{}, nil
}
