package rules

import (
	"errors"
	"fmt"
	"sort"

	"github.com/unprofession-al/bpmon/status"
)

type Rules map[int]Rule

type Rule struct {
	Must       []string      `yaml:"must"`
	MustNot    []string      `yaml:"must_not"`
	Then       string        `yaml:"then"`
	ThenStatus status.Status `yaml:"-"`
}

func (rules Rules) Analyze(values map[string]bool) (status.Status, error) {
	var order []int
	for index := range rules {
		order = append(order, index)
	}
	sort.Ints(order)

	for _, index := range order {
		matchMustCond := true
		matchMustNotCond := true
		rule := rules[index]

		for _, keyname := range rule.Must {
			if val, ok := values[keyname]; ok {
				if !val {
					matchMustCond = false
					break
				}
			} else {
				return status.Unknown, errors.New(fmt.Sprintf("Key '%s' from rule with order %d does not exist", keyname, index))
			}
		}

		for _, keyname := range rule.MustNot {
			if val, ok := values[keyname]; ok {
				if val {
					matchMustNotCond = false
					break
				}
			} else {
				return status.Unknown, errors.New(fmt.Sprintf("Key '%s' from rule with order %d does not exist", keyname, index))
			}
		}

		if matchMustCond && matchMustNotCond {
			return rule.ThenStatus, nil
		}
	}
	return status.Unknown, errors.New("No rule matched")
}
