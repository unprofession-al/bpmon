package status

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mgutz/ansi"
)

type Status int

const (
	Ok = iota
	Nok
	Unknown
)

const (
	OkString      = "ok"
	NokString     = "not ok"
	UnknownString = "unknown"
)

func (s Status) String() string {
	var out string
	switch s {
	case Ok:
		out = OkString
	case Nok:
		out = NokString
	case Unknown:
		out = UnknownString
	}
	return out
}

func FromString(in string) (Status, error) {
	switch strings.ToLower(in) {
	case OkString:
		return Ok, nil
	case NokString:
		return Nok, nil
	case UnknownString:
		return Unknown, nil
	default:
		return Unknown, errors.New(fmt.Sprintf("String '%s' is not a valid status", in))
	}
}

func (s Status) ToInt() int {
	return int(s)
}

func (s Status) Colorize(in string) string {
	var out string
	switch s {
	case Ok:
		out = ansi.Color(in, "green")
	case Nok:
		out = ansi.Color(in, "red+b")
	case Unknown:
		out = ansi.Color(in, "cyan+b")
	}
	return out
}

func BoolAsStatus(ok bool) Status {
	if ok {
		return Ok
	} else {
		return Nok
	}
}

func (s Status) ToBool() bool {
	if s == Nok {
		return false
	}
	return true
}

func (s *Status) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux string
	if err := unmarshal(&aux); err != nil {
		return err
	}
	st, err := FromString(aux)
	s = &st
	return err
}

func (s Status) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}
