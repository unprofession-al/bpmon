package bpmon

import (
	"time"

	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

type StoreMock struct{}

func (s StoreMock) Write(p *store.ResultSet) error {
	return nil
}

func (s StoreMock) Health() (string, error) {
	return "all fine", nil
}

func (s StoreMock) GetLatest(rs store.ResultSet) (store.ResultSet, error) {
	return store.ResultSet{}, nil
}

func (s StoreMock) GetSpans(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Span, error) {
	return []store.Span{}, nil
}

func (s StoreMock) Annotate(id store.ID, annotation string) (store.ResultSet, error) {
	return store.ResultSet{}, nil
}
