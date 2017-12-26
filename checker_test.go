package bpmon

import (
	"errors"
	"time"

	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
)

type CheckerMock struct{}

func (chk CheckerMock) Status(host, service string) (time.Time, string, map[string]bool, error) {
	vals := map[string]bool{
		"good":    false,
		"bad":     false,
		"error":   false,
		"unknown": false,
	}

	now := time.Now()

	switch service {
	case "good":
		vals["good"] = true
		return now, "good", vals, nil
	case "bad":
		vals["bad"] = true
		return now, "bad", vals, nil
	case "error":
		vals["error"] = true
		return now, "error", vals, errors.New("Error occured")
	default:
		vals["unknown"] = true
		return now, "unknown", vals, nil
	}
}

func (chk CheckerMock) Values() []string {
	return []string{"good", "bad", "unknown", "error"}
}

func (chk CheckerMock) DefaultRules() rules.Rules {
	rules := rules.Rules{
		10: rules.Rule{
			Must:    []string{"bad"},
			MustNot: []string{},
			Then:    status.Nok,
		},
		20: rules.Rule{
			Must:    []string{"unknown"},
			MustNot: []string{},
			Then:    status.Unknown,
		},
		9999: rules.Rule{
			Must:    []string{},
			MustNot: []string{},
			Then:    status.Ok,
		},
	}
	return rules
}
