package bpmon

import (
	"errors"
	"time"

	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
)

type CheckerMock struct{}

func (chk CheckerMock) Status(host, service string) checker.Result {
	out := checker.Result{
		Timestamp: time.Now()
	}

	out.Values = map[string]bool{
		"good":    false,
		"bad":     false,
		"error":   false,
		"unknown": false,
	}

	switch service {
	case "good":
		out.Values["good"] = true
	case "bad":
		out.Values["bad"] = true
	case "error":
		out.Values["error"] = true
		out.Error = errors.New("Error occured")
	default:
		out.Values["unknown"] = true
	}
	return out
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
