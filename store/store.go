// Package store provides an interface that allows to implement various
// backends to be loaded in compile time. This makes the persistence layer
// of BPMON interchangeable.
package store

import (
	"errors"
	"sync"
	"time"

	"github.com/unprofession-al/bpmon/status"
)

// Conf keeps the counfiguration for a store implementation. This struct
// is passed to the store implementatinon itself via the registerd setup
// function. The field 'Kind' is used to determine which provider is
// requested.
type Conf struct {
	Kind          string        `yaml:"kind"`
	Connection    string        `yaml:"connection"`
	Timeout       time.Duration `yaml:"timeout"`
	SaveOK        []string      `yaml:"save_ok"`
	GetLastStatus bool          `yaml:"get_last_status"`
	Debug         bool          `yaml:"debug"`
}

var (
	sMu sync.Mutex
	s   = make(map[string]func(Conf) (Accessor, error))
)

// Register must be called in the init function of each store implementation.
// The Register function will panic if two store impelmentations with the same
// name try to register themselfs.
func Register(name string, setupFunc func(Conf) (Accessor, error)) {
	sMu.Lock()
	defer sMu.Unlock()
	if _, dup := s[name]; dup {
		panic("store: Register called twice for store " + name)
	}
	s[name] = setupFunc
}

// New well return a configured instance of a store implementation. The
// implementation requested is determined by the 'Kind' field of the
// configuration struct.
func New(conf Conf) (Accessor, error) {
	setupFunc, ok := s[conf.Kind]
	if !ok {
		return nil, errors.New("store: store '" + conf.Kind + "' does not exist")
	}
	return setupFunc(conf)
}

// Accessor is the interface that describes all operations exposed by a store.
type Accessor interface {
	// Write takes a (nested) ResultSet and persists all values (including all
	// child values) to the store.
	Write(input *ResultSet) error

	// GetSpans fetches all time spans (periods where the status of on entity
	// remains the same) between 'start' and 'end'.
	//
	// To determine which spans should be queried, the 'Tags' of the 'ResultSet'
	// provided are considered.
	//
	// If a span cannot be determinded because of a 'StatusChanged' flag, a
	// potential interval is required to _assume_ if a gap between to
	// measurements represents a status change or should be considered a status
	// change to 'status.OK'. This interval should be sightly larger than the
	// interval of your execution interval of 'bpmon write' in order to bp as
	// accurate as possible.
	//
	// Also the results can be filtered by their status. If no status list is
	// provided all spans will be returned.
	GetSpans(input ResultSet, start time.Time, end time.Time, interval time.Duration, statusRequested []status.Status) ([]Span, error)

	// GetEvents fetches all Events (check results that do have the same status
	// as before) between 'start' and 'end'.
	//
	// If an event cannot be determinded because of a 'StatusChanged' flag, a
	// potential interval is required to _assume_ if a gap between to
	// measurements represents a status change or should be considered a status
	// change to 'status.OK'. This interval should be sightly larger than the
	// interval of your execution interval of 'bpmon write' in order to bp as
	// accurate as possible.
	//
	// Also the results can be filtered by their status. If no status list is
	// provided all events will be returned.
	GetEvents(start time.Time, end time.Time, interval time.Duration, statusRequested []status.Status) ([]Event, error)

	// GetLatest returns a representation of the latest persisted ResultSet
	// matching the 'Tags' of the 'ResultSet' provided as input.
	GetLatest(input ResultSet) (ResultSet, error)

	// AnnotateEvent persists an annotation string on the event described via
	// its 'ID'. It also updates its field 'Annotated' to 'true'.
	AnnotateEvent(id ID, annotation string) (ResultSet, error)
}
