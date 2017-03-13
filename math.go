package bpmon

import (
	"errors"
	"strconv"
	"strings"
)

func and(values []bool) bool {
	for _, val := range values {
		if val == false {
			return false
		}
	}
	return true
}

func or(values []bool) bool {
	if len(values) < 1 {
		return true
	}
	for _, val := range values {
		if val == true {
			return true
		}
	}
	return false
}

func min(values []bool, min float64) bool {
	count := 0.0
	for _, val := range values {
		if val == true {
			count++
		}
		if count >= min {
			return true
		}
	}
	if count >= min {
		return true
	}
	return false
}

func minpercent(values []bool, minpercent float64) bool {
	if len(values) < 1 {
		return true
	}
	max := float64(len(values))
	target := max * minpercent / 100.0
	count := 0.0
	for _, val := range values {
		if val == true {
			count++
		}
		if count >= target {
			return true
		}
	}
	return false
}

func calculate(operation string, values []bool) (bool, error) {
	var out bool

	op, arg, err := parseOp(operation)
	if err != nil {
		return out, err
	}

	switch op {
	case "and":
		out = and(values)
	case "or":
		out = or(values)
	case "min":
		out = min(values, arg)
	case "minpercent":
		out = minpercent(values, arg)
	default:
		return true, errors.New("Operation '" + operation + "' unknown")
	}

	return out, nil
}

func parseOp(operation string) (op string, arg float64, err error) {
	opParts := strings.Split(strings.ToLower(operation), " ")
	op = opParts[0]

	if len(opParts) > 1 {
		arg, err = strconv.ParseFloat(opParts[1], 64)
		if err != nil {
			return "", 0.0, err
		}
	}
	return op, arg, err
}
