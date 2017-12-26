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

func Register(name string, setupFunc func(Conf) (Checker, error)) {
	cMu.Lock()
	defer cMu.Unlock()
	if _, dup := c[name]; dup {
		panic("checker: Register called twice for store " + name)
	}
	c[name] = setupFunc
}

func New(conf Conf) (Checker, error) {
	setupFunc, ok := c[conf.Kind]
	if !ok {
		return nil, errors.New("checker: checker '" + conf.Kind + "' does not exist")
	}
	return setupFunc(conf)
}

type Conf struct {
	Kind       string `yaml:"kind"`
	Connection string `yaml:"connection"`
}

type Checker interface {
	Status(string, string) (time.Time, string, map[string]bool, error)
	Values() []string
	DefaultRules() rules.Rules
}
