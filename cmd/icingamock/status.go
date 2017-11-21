package main

import (
	"fmt"
	"strings"
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
		return Unknown, fmt.Errorf("String '%s' is not a valid status", in)
	}
}

func FromInt64(in int64) (Status, error) {
	switch in {
	case Ok:
		return Ok, nil
	case Nok:
		return Nok, nil
	case Unknown:
		return Unknown, nil
	default:
		return Unknown, fmt.Errorf("Integer '%d' is not a valid status", in)
	}
}

func (s Status) Int() int {
	return int(s)
}

func FromBool(ok bool) Status {
	if ok {
		return Ok
	} else {
		return Nok
	}
}

func (s Status) Bool() bool {
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
