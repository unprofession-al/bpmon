package bpmon

import (
	"errors"
	"fmt"
	"sort"

	"github.com/unprofession-al/bpmon/status"
)

type Rules map[int]Rule

type Rule struct {
	Must       []string
	MustNot    []string
	Then       string
	thenStatus status.Status
}

func (rules Rules) Analyze(svc SvcResult) (status.Status, error) {
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
			if val, ok := svc.Vals[keyname]; ok {
				if !val {
					matchMustCond = false
					break
				}
			} else {
				return status.Unknown, errors.New(fmt.Sprintf("Key '%s' from rule with order %d does not exist", keyname, index))
			}
		}

		for _, keyname := range rule.MustNot {
			if val, ok := svc.Vals[keyname]; ok {
				if val {
					matchMustNotCond = false
					break
				}
			} else {
				return status.Unknown, errors.New(fmt.Sprintf("Key '%s' from rule with order %d does not exist", keyname, index))
			}
		}

		if matchMustCond && matchMustNotCond {
			return rule.thenStatus, nil
		}
	}
	return status.Unknown, errors.New("No rule matched")
}
