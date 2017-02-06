package cmd

import "github.com/mgutz/ansi"

type Status int

const (
	StatusOK = iota
	StatusNOK
	StatusUnknown
)

func (s Status) String() string {
	var out string
	switch s {
	case StatusOK:
		out = "ok"
	case StatusNOK:
		out = "not ok"
	case StatusUnknown:
		out = "unknown"
	}
	return out
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
