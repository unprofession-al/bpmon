package rules

import (
	"reflect"
	"testing"

	"github.com/unprofession-al/bpmon/status"
)

var testRules = map[string]Rules{
	"base": Rules{
		10: Rule{
			Must:    []string{"bad"},
			MustNot: []string{},
			Then:    status.NOK,
		},
		11: Rule{
			Must:    []string{},
			MustNot: []string{"known"},
			Then:    status.Unknown,
		},
		20: Rule{
			Must:    []string{},
			MustNot: []string{},
			Then:    status.OK,
		},
	},
	"additional": Rules{
		15: Rule{
			Must:    []string{"unknown"},
			MustNot: []string{},
			Then:    status.Unknown,
		},
	},
	"base+additional": Rules{
		10: Rule{
			Must:    []string{"bad"},
			MustNot: []string{},
			Then:    status.NOK,
		},
		11: Rule{
			Must:    []string{},
			MustNot: []string{"known"},
			Then:    status.Unknown,
		},
		15: Rule{
			Must:    []string{"unknown"},
			MustNot: []string{},
			Then:    status.Unknown,
		},
		20: Rule{
			Must:    []string{},
			MustNot: []string{},
			Then:    status.OK,
		},
	},
	"overwrite": Rules{
		10: Rule{
			Must:    []string{"the worst"},
			MustNot: []string{},
			Then:    status.Unknown,
		},
	},
	"base+overwrite": Rules{
		10: Rule{
			Must:    []string{"the worst"},
			MustNot: []string{},
			Then:    status.Unknown,
		},
		11: Rule{
			Must:    []string{},
			MustNot: []string{"known"},
			Then:    status.Unknown,
		},
		20: Rule{
			Must:    []string{},
			MustNot: []string{},
			Then:    status.OK,
		},
	},
}

func TestRuleMergingAdd(t *testing.T) {
	rules := Rules{}
	for k, v := range testRules["base"] {
		rules[k] = v
	}
	rules.Merge(testRules["additional"])
	eq := reflect.DeepEqual(rules, testRules["base+additional"])
	if !eq {
		t.Errorf("Results do not match: '%v' vs. '%v'", rules, testRules["base+additional"])
	}
}

func TestRuleMergingOverwrite(t *testing.T) {
	rules := Rules{}
	for k, v := range testRules["base"] {
		rules[k] = v
	}
	rules.Merge(testRules["overwrite"])
	eq := reflect.DeepEqual(rules, testRules["base+overwrite"])
	if !eq {
		t.Errorf("Results do not match: '%v' vs. '%v'", rules, testRules["base+overwrite"])
	}
}

func TestRuleAnalyze(t *testing.T) {
	testsets := map[string]struct {
		desc        string
		test        map[string]bool
		status      status.Status
		errExpected bool
	}{
		"good and not bad": {
			test: map[string]bool{
				"good":  true,
				"bad":   false,
				"known": true,
			},
			status:      status.OK,
			errExpected: false,
		},
		"must fail because key 'bad' and 'known' do not exist": {
			test: map[string]bool{
				"good": true,
			},
			status:      status.Unknown,
			errExpected: true,
		},
		"unknown": {
			test: map[string]bool{
				"good":  false,
				"bad":   false,
				"known": false,
			},
			status:      status.Unknown,
			errExpected: false,
		},
	}

	rules := testRules["base"]

	for name, ts := range testsets {
		s, err := rules.Analyze(ts.test)
		if ts.errExpected && err == nil {
			t.Errorf("Error expected but got nil")
		} else if !ts.errExpected && err != nil {
			t.Errorf("No error expected for test '%s' but got error: %s", name, err.Error())
		}
		if s != ts.status {
			t.Errorf("Expected status to be '%s', got '%s' for %s", ts.status, s, name)
		}
	}
}
