package bpmon

import (
	"time"

	"github.com/unprofession-al/bpmon/persistence"
)

type PersistenceMock struct{}

func (pp PersistenceMock) Write(p *persistence.ResultSet) error {
	return nil
}

func (pp PersistenceMock) GetLatest(rs persistence.ResultSet) (persistence.ResultSet, error) {
	return persistence.ResultSet{}, nil
}

func (pp PersistenceMock) GetEvents(rs persistence.ResultSet, start time.Time, end time.Time, interval time.Duration) ([]persistence.Event, error) {
	return []persistence.Event{}, nil
}
