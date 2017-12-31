package influx

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

func (i Influx) GetEvents(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Event, error) {
	var out []store.Event
	var err error
	if i.getLastStatus {
		for _, saveOkKind := range i.saveOK {
			if saveOkKind == rs.Kind() {
				out, err = i.getEvents(rs, start, end)
				return filterByStatus(out, stati), err
			}
		}
	}
	out, err = i.assumeEvents(rs, start, end, interval)
	return filterByStatus(out, stati), err
}

func filterByStatus(in []store.Event, stati []status.Status) []store.Event {
	if len(stati) == 0 {
		return in
	}
	var out []store.Event
	for _, event := range in {
		for _, st := range stati {
			if event.Status == st {
				out = append(out, event)
			}
		}
	}
	return out
}

func (i Influx) getEvents(rs store.ResultSet, start time.Time, end time.Time) ([]store.Event, error) {
	out := []store.Event{}
	totalDuration := end.Sub(start).Seconds()

	query := NewSelectQuery().From(rs.Kind()).Between(start, end).FilterTags(rs.Tags).Filter("changed = true")
	resultsets, err := i.Run(query)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return out, errors.New(msg)
	}

	earliestEvent := end
	for i, resultset := range resultsets {
		current := store.Event{
			Tags:       resultset.Tags,
			Pseudo:     false,
			Status:     resultset.Status,
			Annotation: resultset.Annotation,
		}
		current.Start = resultset.Start
		if current.Start.Before(earliestEvent) {
			earliestEvent = current.Start
		}
		current.End = end
		if next := i + 1; next < len(resultsets) {
			current.End = resultsets[i+1].Start
		}
		current.Duration = current.End.Sub(current.Start).Seconds()
		current.DurationPercent = 100.0 / float64(totalDuration) * float64(current.Duration)
		current.SetID()
		out = append(out, current)
	}

	// get last state before the time window specified by 'start' and 'end'
	gap, _ := time.ParseDuration("30m")
	query = NewSelectQuery().From(rs.Kind()).Between(end.Add(gap*-1), end).FilterTags(rs.Tags).OrderBy("time").Desc().Limit(1)
	last, err := i.First(query)

	if err != nil {
		// if no state at all is found
		complete := store.Event{
			Status:          status.Unknown,
			Pseudo:          true,
			Annotation:      "no such data found",
			Duration:        totalDuration,
			DurationPercent: 100.0,
			Start:           start,
			End:             end,
			Tags:            rs.Tags,
		}
		complete.SetID()
		out = append([]store.Event{complete}, out...)
	} else {
		duration := earliestEvent.Sub(start).Seconds()
		durationPercent := 100.0 / float64(totalDuration) * float64(duration)
		first := store.Event{
			Start:           start,
			End:             earliestEvent,
			Pseudo:          true,
			Duration:        duration,
			DurationPercent: durationPercent,
			Tags:            last.Tags,
			Status:          last.Status,
			Annotation:      last.Annotation,
		}
		first.SetID()
		out = append([]store.Event{first}, out...)
	}

	return out, nil
}

func (i Influx) assumeEvents(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration) ([]store.Event, error) {
	duration := end.Sub(start).Seconds()
	events := []store.Event{
		store.Event{
			Status:     status.Ok,
			Annotation: "",
			Start:      start,
			End:        end,
			Tags:       rs.Tags,
		},
	}

	query := NewSelectQuery().From(rs.Kind()).Between(start, end).FilterTags(rs.Tags)
	rows, err := i.Run(query)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return events, errors.New(msg)
	}

	lastIndex := len(events) - 1

	for _, row := range rows {
		last := events[lastIndex]
		replace := false

		current := store.Event{
			Tags:       row.Tags,
			Status:     row.Status,
			Start:      row.Start,
			Annotation: row.Annotation,
		}
		current.End = current.Start.Add(interval)
		if current.Start.Before(last.End) {
			if last.Status == current.Status {
				current.Start = last.Start
				current.Annotation = last.Annotation
				replace = true
			} else {
				last.End = current.Start
				events[lastIndex] = last
			}
		} else if current.Start.After(last.End) {
			filler := store.Event{
				Start:      last.End,
				End:        current.Start,
				Status:     status.Ok,
				Annotation: "",
			}
			events = append(events, filler)
		}

		// igrnore that case for now
		// if current.End.Before(last.End) {}

		if replace {
			events[lastIndex] = current
		} else {
			events = append(events, current)
		}

		lastIndex = len(events) - 1
	}

	lastEvent := events[len(events)-1]
	if lastEvent.End != end {
		filler := store.Event{
			Start:      lastEvent.End,
			End:        end,
			Status:     status.Ok,
			Annotation: "",
		}
		events = append(events, filler)
	}

	for i, e := range events {
		e.Duration = e.End.Sub(e.Start).Seconds()
		e.DurationPercent = 100.0 / float64(duration) * e.Duration
		e.SetID()
		events[i] = e
	}
	return events, nil
}

func (i Influx) AnnotateEvent(id string, annotation string) (store.ResultSet, error) {
	rs, err := i.idToResultSet(id)
	if err != nil {
		return rs, err
	}

	filter := fmt.Sprintf("time = %d", rs.Start.UnixNano())
	query := NewSelectQuery().From(rs.Kind()).FilterTags(rs.Tags).Filter(filter).Limit(1)
	rs, err = i.First(query)
	if err != nil {
		return rs, err
	}

	rs.Annotated = true
	rs.Annotation = annotation
	err = i.Write(&rs)

	return rs, err
}

const (
	pairSeparator    = "="
	tagSeparator     = ";"
	timeTagSeparator = " "
)

func (i Influx) idToResultSet(id string) (store.ResultSet, error) {
	rs := store.ResultSet{
		Tags: make(map[string]string),
	}
	data, err := base64.RawURLEncoding.DecodeString(id)
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
