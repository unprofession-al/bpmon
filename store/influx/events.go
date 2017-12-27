package influx

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

func (i Influx) GetEvents(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Event, error) {
	if i.getLastStatus {
		for _, saveOkKind := range i.saveOK {
			if saveOkKind == rs.Kind() {
				return i.getEvents(rs, start, end, stati)
			}
		}
	}
	return i.assumeEvents(rs, start, end, interval, stati)
}

func (i Influx) getEvents(rs store.ResultSet, start time.Time, end time.Time, stati []status.Status) ([]store.Event, error) {
	out := []store.Event{}
	totalDuration := end.Sub(start).Seconds()

	where := []string{}

	statusWhere := []string{}
	for _, st := range stati {
		statusWhere = append(statusWhere, fmt.Sprintf("status = %d", st.Int()))
	}
	if len(statusWhere) > 0 {
		token := fmt.Sprintf(" ( %s ) ", strings.Join(statusWhere, " OR "))
		where = append(where, token)
	}

	for key, value := range rs.Tags {
		where = append(where, fmt.Sprintf("%s = '%s'", key, value))
	}
	where = append(where, fmt.Sprintf("time < %d", getInfluxTimestamp(end)))
	where = append(where, fmt.Sprintf("time > %d", getInfluxTimestamp(start)))
	where = append(where, "changed = true")

	fields := []string{"time", "status", "annotation"}
	for tag, _ := range rs.Tags {
		fields = append(fields, tag)
	}

	rows, err := i.getAll(fields, rs.Kind(), where, "")
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return out, errors.New(msg)
	}

	earliestEvent := end
	for i, row := range rows {
		current := store.Event{Tags: make(map[string]string)}
		current.Start, err = time.Parse(time.RFC3339, row["time"].(string))
		if err != nil {
			return out, err
		}
		if current.Start.Before(earliestEvent) {
			earliestEvent = current.Start
		}
		current.End = end
		next := i + 1
		if next < len(rows) {
			current.End, err = time.Parse(time.RFC3339, rows[i+1]["time"].(string))
			if err != nil {
				return out, err
			}
		}

		current.Duration = current.End.Sub(current.Start).Seconds()
		current.DurationPercent = 100.0 / float64(totalDuration) * float64(current.Duration)
		statusNumber, err := row["status"].(json.Number).Int64()
		if err != nil {
			return out, err
		}
		current.Status, err = status.FromInt64(statusNumber)
		if err != nil {
			return out, err
		}
		if row["annotation"] != nil {
			current.Annotation = row["annotation"].(string)
		} else {
			current.Annotation = ""
		}

		for tag, _ := range rs.Tags {
			current.Tags[tag] = row[tag].(string)
		}

		current.SetID()

		out = append(out, current)
	}

	// get last state before the time window specified by 'start' and 'end'
	whereLast := []string{}
	for key, value := range rs.Tags {
		whereLast = append(whereLast, fmt.Sprintf("%s = '%s'", key, value))
	}
	whereLast = append(whereLast, fmt.Sprintf("time < %d", getInfluxTimestamp(end)))
	whereLast = append(whereLast, fmt.Sprintf("time > %d", getInfluxTimestamp(start)))
	additional := "ORDER BY time DESC LIMIT 1"
	last, err := i.getOne(fields, rs.Kind(), whereLast, additional)
	if err != nil {
		// if no state at all is found
		complete := store.Event{
			Status:          status.Unknown,
			Annotation:      "no such data found",
			Duration:        totalDuration,
			DurationPercent: 100.0,
			Start:           start,
			End:             end,
			Tags:            rs.Tags,
		}

		complete.SetID()
		for _, st := range stati {
			if complete.Status == st {
				out = append([]store.Event{complete}, out...)
			}
		}
	} else {
		duration := earliestEvent.Sub(start).Seconds()
		durationPercent := 100.0 / float64(totalDuration) * float64(duration)
		first := store.Event{
			Start:           start,
			End:             earliestEvent,
			Duration:        duration,
			DurationPercent: durationPercent,
			Tags:            make(map[string]string),
		}
		first.SetID()
		statusNumber, err := last["status"].(json.Number).Int64()
		if err != nil {
			return out, err
		}
		first.Status, err = status.FromInt64(statusNumber)
		if err != nil {
			return out, err
		}
		if last["annotation"] != nil {
			first.Annotation = last["annotation"].(string)
		} else {
			first.Annotation = ""
		}
		for tag, _ := range rs.Tags {
			first.Tags[tag] = last[tag].(string)
		}
		for _, st := range stati {
			if first.Status == st {
				out = append([]store.Event{first}, out...)
			}
		}
	}

	return out, nil
}

func (i Influx) assumeEvents(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Event, error) {
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

	where := []string{}

	statusWhere := []string{}
	for _, st := range stati {
		statusWhere = append(statusWhere, fmt.Sprintf("status = %d", st.Int()))
	}
	if len(statusWhere) > 0 {
		token := fmt.Sprintf(" ( %s ) ", strings.Join(statusWhere, " OR "))
		where = append(where, token)
	}
	for key, value := range rs.Tags {
		where = append(where, fmt.Sprintf("%s = '%s'", key, value))
	}
	where = append(where, fmt.Sprintf("time < %d", getInfluxTimestamp(end)))
	where = append(where, fmt.Sprintf("time > %d", getInfluxTimestamp(start)))

	fields := []string{"time", "status", "annotation"}
	for tag, _ := range rs.Tags {
		fields = append(fields, tag)
	}

	rows, err := i.getAll(fields, rs.Kind(), where, "")
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return events, errors.New(msg)
	}

	lastIndex := len(events) - 1

	for _, row := range rows {
		last := events[lastIndex]
		replace := false

		current := store.Event{Tags: make(map[string]string)}
		statusNumber, err := row["status"].(json.Number).Int64()
		if err != nil {
			return events, err
		}
		current.Status, err = status.FromInt64(statusNumber)
		if err != nil {
			return events, err
		}
		current.Start, err = time.Parse(time.RFC3339, row["time"].(string))
		if err != nil {
			return events, err
		}
		current.End = current.Start.Add(interval)
		if row["annotation"] != nil {
			current.Annotation = row["annotation"].(string)
		} else {
			current.Annotation = ""
		}

		for tag, _ := range rs.Tags {
			current.Tags[tag] = row[tag].(string)
		}

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
			for _, st := range stati {
				if filler.Status == st {
					events = append(events, filler)
				}
			}
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
		for _, st := range stati {
			if filler.Status == st {
				events = append(events, filler)
			}
		}
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

	where := []string{}
	for key, value := range rs.Tags {
		where = append(where, fmt.Sprintf("%s = '%s'", key, value))
	}
	where = append(where, fmt.Sprintf("time = %d", getInfluxTimestamp(rs.At)))

	point, err := i.getOne([]string{"*"}, rs.Kind(), where, "")
	if err != nil {
		return rs, err
	}

	rs, err = i.asResultSet(point)
	if err != nil {
		return rs, err
	}

	rs, err = i.asResultSet(point)
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

	rs.At = time.Unix(0, nanos)

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
