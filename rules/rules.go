package rules

import (
	"errors"
	"fmt"
	"sort"

	"github.com/unprofession-al/bpmon/status"
)

// Rules is a collection of rules where the key defines the position of the
// of how those rules are applied.
type Rules map[int]Rule

// Rule describes the conditions that must be fulfilled ('Must' and 'MustNot')
// as well as the result of the rule if all conditions are fulfilled ('Then')
type Rule struct {
	// Must is a list of value keys that must be 'true' in order to fulfill
	// the Rule.
	Must []string `yaml:"must"`

	// MustNot is a list of value keys that must be 'false' in order to fulfill
	// the Rule.
	MustNot []string `yaml:"must_not"`

	// Then is the resulting 'Status' if all conditions are fulfilled
	// as defined.
	Then status.Status `yaml:"then"`
}

func (r Rules) Merge(additional Rules) error {
	for order, a := range additional {
		rule := Rule{
			Must:    a.Must,
			MustNot: a.MustNot,
			Then:    a.Then,
		}
		r[order] = rule
	}
	return nil
}

func (r Rules) Analyze(values map[string]bool) (status.Status, error) {
	var order []int
	for index := range r {
		order = append(order, index)
	}
	sort.Ints(order)

	for _, index := range order {
		matchMustCond := true
		matchMustNotCond := true
		rule := r[index]

		for _, keyname := range rule.Must {
			if val, ok := values[keyname]; ok {
				if !val {
					matchMustCond = false
					break
				}
			} else {
				return status.StatusUnknown, fmt.Errorf("Key '%s' from rule with order %d does not exist", keyname, index)
			}
		}

		for _, keyname := range rule.MustNot {
			if val, ok := values[keyname]; ok {
				if val {
					matchMustNotCond = false
					break
				}
			} else {
				return status.StatusUnknown, fmt.Errorf("Key '%s' from rule with order %d does not exist", keyname, index)
			}
		}

		if matchMustCond && matchMustNotCond {
			return rule.Then, nil
		}
	}
	return status.StatusUnknown, errors.New("No rule matched")
}
