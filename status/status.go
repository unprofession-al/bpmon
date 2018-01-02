// Package status provides type status used in BPMON to represent all possible
// stauts as well as some helpers.
package status

import (
	"fmt"

	"github.com/mgutz/ansi"
)

// Status represests the status itself.
type Status int

// The status code list.
const (
	OK = iota
	NOK
	Unknown
)

var statusText = map[int]string{
	OK:      "ok",
	NOK:     "not ok",
	Unknown: "unknown",
}

// String implements the stringer interface.
func (s Status) String() string {
	return statusText[int(s)]
}

// FromString returns a status matching the string provided. If the string does
// not match any status, 'Unknown' as well as an error are returned.
func FromString(in string) (Status, error) {
	for status, text := range statusText {
		if text == in {
			return Status(status), nil
		}
	}
	return Unknown, fmt.Errorf("String '%s' is not a valid status", in)
}

// FromInt64 returns a status matching the int64 provided. If the input does
// not match any status, 'Unknown' as well as an error are returned.
func FromInt64(in int64) (Status, error) {
	switch in {
	case OK:
		return OK, nil
	case NOK:
		return NOK, nil
	case Unknown:
		return Unknown, nil
	default:
		return Unknown, fmt.Errorf("Integer '%d' is not a valid status", in)
	}
}

// Int returns on integer representation of the status.
func (s Status) Int() int {
	return int(s)
}

// Colorize returns a xterm-colorized string of the status.
func (s Status) Colorize(in string) string {
	var out string
	switch s {
	case OK:
		out = ansi.Color(in, "green")
	case NOK:
		out = ansi.Color(in, "red+b")
	case Unknown:
		out = ansi.Color(in, "cyan+b")
	}
	return out
}

// FromBool return a status 'OK' if the bool is true, status 'NOK' if the bool
// is false.
func FromBool(ok bool) Status {
	if ok {
		return OK
	}
	return NOK
}

// Bool returns on boolean representation of the status. Status 'Unknown' is
// considered true.
func (s Status) Bool() bool {
	return s != NOK
}

// UnmarshalYAML implements the Unmarshaler interface of package yaml.
// https://godoc.org/gopkg.in/yaml.v2#Unmarshaler
func (s *Status) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux string
	if err := unmarshal(&aux); err != nil {
		return err
	}
	st, err := FromString(aux)
	s = &st
	return err
}

// MarshalYAML implements the Marshaler interface of package yaml.
// https://godoc.org/gopkg.in/yaml.v2#Marshaler
func (s Status) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}
