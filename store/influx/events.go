package influx

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

func (i Influx) GetEvents(start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Event, error) {
	var out []store.Event

	var statusQuerySegment []string
	for _, st := range stati {
		statusQuerySegment = append(statusQuerySegment, fmt.Sprintf("status = %d", st.Int()))
	}
	sqs := strings.Join(statusQuerySegment, " OR ")

	q := newSelectQuery().From("BP").Between(start, end).Filter("changed = true").Filter(sqs).Filter("annotated = false")
	resultsets, err := i.Run(q)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return out, errors.New(msg)
	}

	for _, rs := range resultsets {
		e := store.Event{
			Status:     rs.Status,
			Annotation: "-",
			Time:       rs.Start,
			Tags:       rs.Tags,
		}

		e.SetID()
		out = append(out, e)
	}

	return []store.Event(out), nil
}

func (i Influx) AnnotateEvent(id store.ID, annotation string) (store.ResultSet, error) {
	rs, err := id.GetResultSet()
	if err != nil {
		return rs, err
	}

	filter := fmt.Sprintf("time = %d", rs.Start.UnixNano())
	q := newSelectQuery().From(rs.Kind().String()).FilterTags(rs.Tags).Filter(filter).Limit(1)
	rs, err = i.First(q)
	if err != nil {
		return rs, err
	}

	rs.Annotated = true
	rs.Annotation = annotation
	err = i.Write(&rs)

	return rs, err
}
