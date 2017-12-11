package bpmon

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/unprofession-al/bpmon/status"
)

type Event struct {
	Status          status.Status `json:"status"`
	Annotation      string        `json:"annotation"`
	Start           time.Time     `json:"start"`
	End             time.Time     `json:"end"`
	Duration        float64       `json:"duration"`
	DurationPercent float64       `json:"duration_percent"`
}

func (i Influx) GetEvents(spec map[string]string, start time.Time, end time.Time) ([]Point, error) {
	kind := getKind(spec)
	out := []Point{}
	startTs := getInfluxTimestamp(start)
	endTs := getInfluxTimestamp(end)
	duration := end.Sub(start).Seconds()

	where := ""
	for key, value := range spec {
		if where != "" {
			where = fmt.Sprintf("%s AND ", where)
		}
		where = fmt.Sprintf("%s%s = '%s'", where, key, value)
	}

	q := fmt.Sprintf("SELECT time, status, annotation FROM %s WHERE %s AND changed = true AND time < %d AND time > %d", kind, where, endTs, startTs)
	res, err := queryDB(i.cli, i.database, q)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query `%s`, error is: %s", q, err.Error())
		return out, errors.New(msg)
	}
	earliestEvent := end
	if len(res) >= 1 &&
		len(res[0].Series) >= 1 &&
		len(res[0].Series[0].Values) >= 1 {
		vals := res[0].Series[0].Values

		for i, row := range vals {
			t, err := time.Parse(time.RFC3339, row[0].(string))
			if err != nil {
				return out, err
			}
			if t.Before(earliestEvent) {
				earliestEvent = t
			}
			tEnd := end
			next := i + 1
			if next < len(vals) {
				tEnd, err = time.Parse(time.RFC3339, vals[i+1][0].(string))
				if err != nil {
					return out, err
				}
			}

			eventDuration := tEnd.Sub(t).Seconds()
			eventDurationPercent := 100.0 / float64(duration) * float64(eventDuration)

			fields := make(map[string]interface{})
			fields["status"] = row[1]
			fields["annotation"] = row[2]
			fields["duration"] = eventDuration
			fields["duration_percent"] = eventDurationPercent

			point := Point{
				Timestamp: t,
				Series:    res[0].Series[0].Name,
				Fields:    fields,
			}

			out = append(out, point)
		}

	}

	// get last state before the time window specified by 'start' and 'end'
	q = fmt.Sprintf("SELECT time, status, annotation FROM %s WHERE %s AND time < %d ORDER by time DESC LIMIT 1", kind, where, getInfluxTimestamp(earliestEvent))
	res, err = queryDB(i.cli, i.database, q)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query `%s`, error is: %s", q, err.Error())
		return out, errors.New(msg)
	}
	if len(res) >= 1 &&
		len(res[0].Series) >= 1 &&
		len(res[0].Series[0].Values) >= 1 {
		last := res[0].Series[0].Values[0]
		tEnd := earliestEvent

		eventDuration := tEnd.Sub(start).Seconds()
		eventDurationPercent := 100.0 / float64(duration) * float64(eventDuration)

		fields := make(map[string]interface{})
		fields["status"] = last[1]
		fields["annotation"] = last[2]
		fields["duration"] = eventDuration
		fields["duration_percent"] = eventDurationPercent

		point := Point{
			Timestamp: start,
			Series:    res[0].Series[0].Name,
			Fields:    fields,
		}

		out = append([]Point{point}, out...)
	} else {
		// if no state at all is found
		fields := make(map[string]interface{})
		fields["status"] = 9
		fields["annotation"] = "no such data found"
		fields["duration"] = duration
		fields["duration_percent"] = 100.0

		point := Point{
			Timestamp: start,
			Fields:    fields,
		}

		out = append([]Point{point}, out...)
	}

	return out, nil
}

func (i Influx) AssumeEvents(spec map[string]string, start time.Time, end time.Time, interval time.Duration) ([]Point, error) {
	kind := getKind(spec)
	startTs := getInfluxTimestamp(start)
	endTs := getInfluxTimestamp(end)
	duration := end.Sub(start).Seconds()
	events := []Event{
		Event{Status: status.Ok,
			Annotation: "",
			Start:      start,
			End:        end,
		},
	}
	out := []Point{}

	where := ""
	for key, value := range spec {
		if where != "" {
			where = fmt.Sprintf("%s AND ", where)
		}
		where = fmt.Sprintf("%s%s = '%s'", where, key, value)
	}

	q := fmt.Sprintf("SELECT time, status, annotation FROM %s WHERE %s AND time < %d AND time > %d", kind, where, endTs, startTs)
	res, err := queryDB(i.cli, i.database, q)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query `%s`, error is: %s", q, err.Error())
		return out, errors.New(msg)
	}

	if len(res) >= 1 &&
		len(res[0].Series) >= 1 &&
		len(res[0].Series[0].Values) >= 1 {

		vals := res[0].Series[0].Values

		lastIndex := len(events) - 1

		for _, row := range vals {
			last := events[lastIndex]
			replace := false

			current := Event{}
			current.Status, _ = status.FromString(row[1].(json.Number).String())
			current.Start, err = time.Parse(time.RFC3339, row[0].(string))
			if err != nil {
				return out, err
			}
			current.End = current.Start.Add(interval)
			if row[2] != nil {
				current.Annotation = row[2].(string)
			} else {
				current.Annotation = ""
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
				filler := Event{
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
	}

	lastEvent := events[len(events)-1]
	if lastEvent.End != end {
		filler := Event{
			Start:      lastEvent.End,
			End:        end,
			Status:     status.Ok,
			Annotation: "",
		}
		events = append(events, filler)
	}

	for _, ev := range events {
		ev.Duration = ev.End.Sub(ev.Start).Seconds()
		ev.DurationPercent = 100.0 / float64(duration) * ev.Duration

		fields := make(map[string]interface{})
		fields["status"] = ev.Status
		fields["annotation"] = ev.Annotation
		fields["end"] = ev.End
		fields["duration"] = ev.Duration
		fields["duration_percent"] = ev.DurationPercent

		point := Point{
			Timestamp: ev.Start,
			Series:    kind,
			Fields:    fields,
		}

		out = append(out, point)
	}
	return out, nil
}
