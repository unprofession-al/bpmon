package cmd

import "github.com/mgutz/ansi"

type status int

const (
	StatusOK = iota
	StatusNOK
	StatusUnknown
)

func (s status) String() string {
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

func (s status) toInt() int {
	return int(s)
}

func (s status) Colorize(in string) string {
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

func boolAsStatus(ok bool) status {
	if ok {
		return StatusOK
	} else {
		return StatusNOK
	}
}
