package bpmon

import (
	"errors"
	"time"

	"github.com/unprofession-al/bpmon/internal/checker"
	"github.com/unprofession-al/bpmon/internal/rules"
	"github.com/unprofession-al/bpmon/internal/status"
)

type CheckerMock struct{}

func (chk CheckerMock) Status(host, service string) checker.Result {
	out := checker.Result{
		Timestamp: time.Now(),
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
		out.Error = errors.New("Error occurred")
	default:
		out.Values["unknown"] = true
	}
	return out
}

func (chk CheckerMock) Values() []string {
	return []string{"good", "bad", "unknown", "error"}
}

func (chk CheckerMock) Health() (string, error) {
	return "all fine", nil
}

func (chk CheckerMock) DefaultRules() rules.Rules {
	rules := rules.Rules{
		10: rules.Rule{
			Must:    []string{"bad"},
			MustNot: []string{},
			Then:    status.StatusNOK,
		},
		20: rules.Rule{
			Must:    []string{"unknown"},
			MustNot: []string{},
			Then:    status.StatusUnknown,
		},
		9999: rules.Rule{
			Must:    []string{},
			MustNot: []string{},
			Then:    status.StatusOK,
		},
	}
	return rules
}
