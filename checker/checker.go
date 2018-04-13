// Package checker provides a generic interface to interact with service status
// provider implementations such as the Icinga implementaton.
//
// The checker package must be used together with on implementation.
package checker

import (
	"errors"
	"sync"
	"time"

	"github.com/unprofession-al/bpmon/rules"
)

var (
	cMu sync.Mutex
	c   = make(map[string]func(Conf) (Checker, error))
)

// Conf keeps the configuration of the checker implementation. This struct is
// passed to the store implementatinon itself via the registerd setup function.
// The field 'Kind' is used to determine which provider is requested.
type Conf struct {
	Kind          string `yaml:"kind"`
	Connection    string `yaml:"connection"`
	TLSSkipVerify bool   `yaml:"tls_skip_verify"`
}

// Validate returns an list of error messages as well as an error if a configuration
// contains invalid values.
func (c Conf) Validate() ([]string, error) {
	errs := []string{}
	if c.Kind == "" {
		errs = append(errs, "Field 'kind' cannot be empty.")
	}
	if c.Connection == "" {
		errs = append(errs, "Field 'connection' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'checker' has errors")
		return errs, err
	}
	return errs, nil
}

// UnmarshalYAML ensures reasonable defaults.
func (c *Conf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type ConfDefaulted Conf

	var defaults = ConfDefaulted{
		Kind:          "icinga",
		Connection:    "http://127.0.0.1:8765/icinga/_",
		TLSSkipVerify: false,
	}

	out := defaults
	err := unmarshal(&out)
	*c = Conf(out)
	return err
}

// Checker interface needs to be implemented in order to provide a Checker
// backend such as Icinga.
type Checker interface {
	// Health tries to connect to the checker implementation and checks its
	// status.
	Health() (string, error)

	// Status takes a host string as well as a service string and returns
	// 'Result' of the stuct of the check.
	Status(host string, service string) Result

	// Values returns a lists of value names that a 'Result' stuct will contain
	// when 'Status()' is called.
	Values() []string

	// Each checked provides its own default rules on which a 'Result' status
	// is evaluated in order to get a 'status.Status'.
	DefaultRules() rules.Rules
}

// Register must be called in the init function of each checker implementation.
// The Register function will panic if two checker impelmentations with the
// same name try to register themselfs.
func Register(name string, setupFunc func(Conf) (Checker, error)) {
	cMu.Lock()
	defer cMu.Unlock()
	if _, dup := c[name]; dup {
		panic("checker: Register called twice for store " + name)
	}
	c[name] = setupFunc
}

// New well return a configured instance of a checker implementation. The
// implementation requested is determined by the 'Kind' field of the
// configuration struct.
func New(conf Conf) (Checker, error) {
	setupFunc, ok := c[conf.Kind]
	if !ok {
		return nil, errors.New("checker: checker '" + conf.Kind + "' does not exist")
	}
	return setupFunc(conf)
}

// Result is returned on a service status check. It contains all relevant
// information in the effective result of the check in the 'Values' map.
// If an error occures while performing the check, it is stored in the 'Error'
// field.
type Result struct {
	Timestamp time.Time
	Message   string
	Values    map[string]bool
	Error     error
}
