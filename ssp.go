package bpmon

import (
	"time"

	"github.com/unprofession-al/bpmon/rules"
)

type ServiceStatusProvider interface {
	Status(string, string) (time.Time, string, map[string]bool, error)
	Values() []string
	DefaultRules() rules.Rules
}
