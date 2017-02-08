package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mgutz/ansi"
)

type Status int

const (
	StatusOK = iota
	StatusNOK
	StatusUnknown
)

const (
	StatusOKString      = "ok"
	StatusNOKString     = "not ok"
	StatusUnknownString = "unknown"
)

func (s Status) String() string {
	var out string
	switch s {
	case StatusOK:
		out = StatusOKString
	case StatusNOK:
		out = StatusNOKString
	case StatusUnknown:
		out = StatusUnknownString
	}
	return out
}

func StatusFromString(in string) (Status, error) {
	switch strings.ToLower(in) {
	case StatusOKString:
		return StatusOK, nil
	case StatusNOKString:
		return StatusNOK, nil
	case StatusUnknownString:
		return StatusUnknown, nil
	default:
		return StatusUnknown, errors.New(fmt.Sprintf("String '%s' is not a valid status", in))
	}
}

func (s Status) toInt() int {
	return int(s)
}

func (s Status) Colorize(in string) string {
	var out string
	switch s {
	case StatusOK:
		out = ansi.Color(in, "green")
	case StatusNOK:
		out = ansi.Color(in, "red+b")
	case StatusUnknown:
		out = ansi.Color(in, "cyan+b")
	}
	return out
}

func boolAsStatus(ok bool) Status {
	if ok {
		return StatusOK
	} else {
		return StatusNOK
	}
}

func (s Status) toBool() bool {
	if s == StatusNOK {
		return false
	}
	return true
}
