package store

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ID is a unique id per 'Event'/'Span'. In order to bp translatable to a
// ResultSet, the ID is composed of its timestamp as well its tags.
// The composed string is then BASE64 encoded (no security reasons here for
// a solid encryption, therefore BASE64 is fine).
type ID string

const (
	pairSeparator    = "="
	tagSeparator     = ";"
	timeTagSeparator = " "
)

// NewID generates an ID based on the timestamp as well as the tags.
func NewID(timestamp time.Time, tags map[Kind]string) ID {
	var pairs []string
	for key, value := range tags {
		pairs = append(pairs, key.String()+pairSeparator+value)
	}
	s := fmt.Sprintf("%v%s%s", timestamp.UnixNano(), timeTagSeparator, strings.Join(pairs, tagSeparator))
	return ID(base64.RawURLEncoding.EncodeToString([]byte(s)))
}

// GetResultSet returns a ResultSet based on the Infos contained in the
// ID string. If the encoding in gibberish, an error is returned.
func (eid ID) GetResultSet() (ResultSet, error) {
	rs := ResultSet{
		Tags: make(map[Kind]string),
	}
	data, err := base64.RawURLEncoding.DecodeString(string(eid))
	if err != nil {
		return rs, err
	}

	elements := strings.SplitN(string(data), timeTagSeparator, 2)
	if len(elements) != 2 {
		return rs, errors.New("malformed Event ID")
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
			return rs, errors.New("malformed Event ID")
		}
		rs.Tags[Kind(touple[0])] = touple[1]
	}

	return rs, nil
}
