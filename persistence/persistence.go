package persistence

import (
	"errors"
	"sync"
	"time"
)

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
	s   = make(map[string]func(Conf) (Persistence, error))
)

func Register(name string, setupFunc func(Conf) (Persistence, error)) {
	sMu.Lock()
	defer sMu.Unlock()
	if _, dup := s[name]; dup {
		panic("store: Register called twice for store " + name)
	}
	s[name] = setupFunc
}

func New(conf Conf) (Persistence, error) {
	setupFunc, ok := s[conf.Kind]
	if !ok {
		return nil, errors.New("store: store '" + conf.Kind + "' does not exist")
	}
	return setupFunc(conf)
}

type Persistence interface {
	GetEvents(ResultSet, time.Time, time.Time, time.Duration) ([]Event, error)
	GetLatest(ResultSet) (ResultSet, error)
	Write(*ResultSet) error
}
