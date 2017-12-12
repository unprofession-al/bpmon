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

type EventProvider struct {
	PersistenceProvider
	saveOk        []string
	getLastStatus bool
}

func NewEventProvider(pp PersistenceProvider, saveOk []string, getLastStatus bool) EventProvider {
	return EventProvider{pp, saveOk, getLastStatus}
}

func (ep EventProvider) GetEvents(spec map[string]string, start time.Time, end time.Time, interval time.Duration) ([]Event, error) {
	kind := getKind(spec)
	if ep.getLastStatus {
		for _, saveOkKind := range ep.saveOk {
			if saveOkKind == kind {
				return ep.getEvents(spec, start, end)
			}
		}
	}
	return ep.assumeEvents(spec, start, end, interval)
}

func (ep EventProvider) getEvents(spec map[string]string, start time.Time, end time.Time) ([]Event, error) {
	out := []Event{}
	totalDuration := end.Sub(start).Seconds()

	kind := getKind(spec)

	where := []string{}
	for key, value := range spec {
		where = append(where, fmt.Sprintf("%s = '%s'", key, value))
	}
	where = append(where, fmt.Sprintf("time < %d", getInfluxTimestamp(end)))
	where = append(where, fmt.Sprintf("time > %d", getInfluxTimestamp(start)))
	where = append(where, "changed = true")

	fields := []string{"time", "status", "annotation"}

	rows, err := ep.GetAll(fields, kind, where, "")
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return out, errors.New(msg)
	}

	earliestEvent := end
	for i, row := range rows {
		current := Event{}
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

		out = append(out, current)
	}

	// get last state before the time window specified by 'start' and 'end'
	whereLast := []string{}
	for key, value := range spec {
		whereLast = append(whereLast, fmt.Sprintf("%s = '%s'", key, value))
	}
	whereLast = append(whereLast, fmt.Sprintf("time < %d", getInfluxTimestamp(end)))
	whereLast = append(whereLast, fmt.Sprintf("time > %d", getInfluxTimestamp(start)))
	additional := "ORDER BY time DESC LIMIT 1"
	last, err := ep.GetOne(fields, kind, whereLast, additional)
	if err != nil {
		// if no state at all is found
		complete := Event{
			Status:          status.Unknown,
			Annotation:      "no such data found",
			Duration:        totalDuration,
			DurationPercent: 100.0,
			Start:           start,
			End:             end,
		}
		out = append([]Event{complete}, out...)
	} else {
		duration := earliestEvent.Sub(start).Seconds()
		durationPercent := 100.0 / float64(totalDuration) * float64(duration)
		first := Event{
			Start:           start,
			End:             earliestEvent,
			Duration:        duration,
			DurationPercent: durationPercent,
		}
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
		out = append([]Event{first}, out...)
	}

	return out, nil
}

func (ep EventProvider) assumeEvents(spec map[string]string, start time.Time, end time.Time, interval time.Duration) ([]Event, error) {
	duration := end.Sub(start).Seconds()
	events := []Event{
		Event{
			Status:     status.Ok,
			Annotation: "",
			Start:      start,
			End:        end,
		},
	}

	kind := getKind(spec)

	where := []string{}
	for key, value := range spec {
		where = append(where, fmt.Sprintf("%s = '%s'", key, value))
	}
	where = append(where, fmt.Sprintf("time < %d", getInfluxTimestamp(end)))
	where = append(where, fmt.Sprintf("time > %d", getInfluxTimestamp(start)))

	fields := []string{"time", "status", "annotation"}

	rows, err := ep.GetAll(fields, kind, where, "")
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return events, errors.New(msg)
	}

	lastIndex := len(events) - 1

	for _, row := range rows {
		last := events[lastIndex]
		replace := false

		current := Event{}
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

	for i, e := range events {
		e.Duration = e.End.Sub(e.Start).Seconds()
		e.DurationPercent = 100.0 / float64(duration) * e.Duration
		events[i] = e
	}
	return events, nil
}