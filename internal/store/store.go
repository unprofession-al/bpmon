// Package store provides an interface that allows to implement various
// backends to be loaded in compile time. This makes the persistence layer
// of BPMON interchangeable.
package store

import (
	"errors"
	"sync"
	"time"

	"github.com/unprofession-al/bpmon/internal/status"
)

// Config keeps the counfiguration for a store implementation. This struct
// is passed to the store implementatinon itself via the registered setup
// function. The field 'Kind' is used to determine which provider is
// requested.
type Config struct {
	// kind defines the store implementation to be used by BPMON. Currently
	// only influx is implemented.
	Kind string `yaml:"kind"`

	// The connection string describes how to connect to your Influx Database.
	// The string needs to follow the pattern:
	//   [protocol]://[user]:[passwd]@[hostname]:[port]
	Connection string `yaml:"connection"`

	// timeout is read as a go (golang) duration, please refer to
	// https://golang.org/pkg/time/#Duration for a detailed explanation.
	Timeout time.Duration `yaml:"timeout"`

	// save_ok tells BPMON which data points should be persisted if the state is 'ok'.
	// By default 'OK' states aro only saved to InfluxDB if its an BP measurement.
	// That means that 'OK' states for KPIs and SVCs will not be saved for the sake of
	// of storage required. 'OK' states of BPs are saved as 'heart beat' of BPMON.
	SaveOK []string `yaml:"save_ok"`

	// This will tell BPMON to compare the current status against the last
	// status saved in InfluxDB and adds some values to the measurement
	// accordingly. This then allows to generate reports such as 'Tell me
	// only when a status is changed from good to bad'. This only runs against
	// types listed in 'save_ok' since only these are persisted 'correctly'.
	GetLastStatus bool `yaml:"get_last_status"`

	// if debug is set to true all queries generated and executed by bpmon will
	// be logged to stdout.
	Debug bool `yaml:"debug"`

	// BPMON verifies if a https connection is trusted. If you wont to trust a
	// connection with an invalid certificate you have to set this to true
	TLSSkipVerify bool `yaml:"tls_skip_verify"`
}

func Defaults() Config {
	return Config(configDefaults)
}

type ConfigDefaulted Config

var configDefaults = ConfigDefaulted{
	Kind:          "influx",
	Timeout:       time.Duration(10 * time.Second),
	SaveOK:        []string{"BP"},
	GetLastStatus: true,
	Debug:         false,
	TLSSkipVerify: false,
}

func (c Config) Validate() ([]string, error) {
	errs := []string{}
	if c.Kind == "" {
		errs = append(errs, "Field 'kind' cannot be empty.")
	}
	if c.Connection == "" {
		errs = append(errs, "Field 'connection' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'store' has errors")
		return errs, err
	}
	return errs, nil
}

var (
	sMu sync.Mutex
	s   = make(map[string]func(Config) (Accessor, error))
)

// Register must be called in the init function of each store implementation.
// The Register function will panic if two store impelmentations with the same
// name try to register themselves.
func Register(name string, setupFunc func(Config) (Accessor, error)) {
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
func New(c Config) (Accessor, error) {
	setupFunc, ok := s[c.Kind]
	if !ok {
		return nil, errors.New("store: store '" + c.Kind + "' does not exist")
	}
	return setupFunc(c)
}

// Accessor is the interface that describes all operations exposed by a store.
type Accessor interface {
	// Health tries to connect to the store implementation and checks its status.
	Health() (string, error)

	// Write takes a (nested) ResultSet and persists all values (including all
	// child values) to the store.
	Write(input *ResultSet) error

	// GetSpans fetches all time spans (periods where the status of on entity
	// remains the same) between 'start' and 'end'.
	//
	// To determine which spans should be queried, the 'Tags' of the 'ResultSet'
	// provided are considered.
	//
	// If a span cannot be determined because of a 'StatusChanged' flag, a
	// potential interval is required to _assume_ if a gap between to
	// measurements represents a status change or should be considered a status
	// change to 'status.OK'. This interval should be sightly larger than the
	// interval of your execution interval of 'bpmon write' in order to bp as
	// accurate as possible.
	//
	// Also the results can be filtered by their status. If no status list is
	// provided all spans will be returned.
	GetSpans(input ResultSet, start time.Time, end time.Time, interval time.Duration, statusRequested []status.Status) ([]Span, error)

	// GetLatest returns a representation of the latest persisted ResultSet
	// matching the 'Tags' of the 'ResultSet' provided as input.
	GetLatest(input ResultSet) (ResultSet, error)

	// Annotate persists an annotation string on the event described via
	// its 'ID'. It also updates its field 'Annotated' to 'true'.
	Annotate(id ID, annotation string) (ResultSet, error)
}
