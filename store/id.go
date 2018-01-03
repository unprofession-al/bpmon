package store

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ID string

const (
	pairSeparator    = "="
	tagSeparator     = ";"
	timeTagSeparator = " "
)

func NewID(timestamp time.Time, tags map[string]string) ID {
	var pairs []string
	for key, value := range tags {
		pairs = append(pairs, key+pairSeparator+value)
	}
	s := fmt.Sprintf("%v%s%s", timestamp.UnixNano(), timeTagSeparator, strings.Join(pairs, tagSeparator))
	return ID(base64.RawURLEncoding.EncodeToString([]byte(s)))
}

func (eid ID) GetResultSet() (ResultSet, error) {
	rs := ResultSet{
		Tags: make(map[string]string),
	}
	data, err := base64.RawURLEncoding.DecodeString(string(eid))
	if err != nil {
		return rs, err
	}

	elements := strings.SplitN(string(data), timeTagSeparator, 2)
	if len(elements) != 2 {
		return rs, errors.New("Malformed Event ID")
	}

	nanos, err := strconv.ParseInt(elements[0], 10, 64)
	if err != nil {
		return rs, err
	}

	rs.Start = time.Unix(0, nanos)

	tags := strings.Split(elements[1], tagSeparator)
	for _, pair := range tags {
		touple := strings.SplitN(pair, pairSeparator, 2)
		if len(touple) != 2 {
			return rs, errors.New("Malformed Event ID")
		}
		rs.Tags[touple[0]] = touple[1]
	}

	return rs, nil
}
