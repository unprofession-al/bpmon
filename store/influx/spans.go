package influx

import (
	"errors"
	"fmt"
	"time"

	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

type spans []store.Span

func (s spans) FilterByStatus(stati []status.Status) spans {
	if len(stati) == 0 {
		return s
	}
	var out spans
	for _, span := range s {
		for _, st := range stati {
			if span.Status == st {
				out = append(out, span)
			}
		}
	}
	return out
}

func (i Influx) GetSpans(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration, stati []status.Status) ([]store.Span, error) {
	var out spans
	var err error
	if i.getLastStatus {
		for _, saveOkKind := range i.saveOK {
			if saveOkKind == rs.Kind() {
				out, err = i.getEvents(rs, start, end)
				return out.FilterByStatus(stati), err
			}
		}
	}
	out, err = i.assumeEvents(rs, start, end, interval)
	return out.FilterByStatus(stati), err
}

func (i Influx) getEvents(rs store.ResultSet, start time.Time, end time.Time) ([]store.Span, error) {
	out := []store.Span{}
	totalDuration := end.Sub(start).Seconds()

	q := newSelectQuery().From(rs.Kind()).Between(start, end).FilterTags(rs.Tags).Filter("changed = true")
	resultsets, err := i.Run(q)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return out, errors.New(msg)
	}

	earliestSpan := end
	for i, resultset := range resultsets {
		current := store.Span{
			Tags:       resultset.Tags,
			Pseudo:     false,
			Status:     resultset.Status,
			Annotation: resultset.Annotation,
		}
		current.Start = resultset.Start
		if current.Start.Before(earliestSpan) {
			earliestSpan = current.Start
		}
		current.End = end
		if next := i + 1; next < len(resultsets) {
			current.End = resultsets[i+1].Start
		}
		current.Duration = current.End.Sub(current.Start).Seconds()
		current.DurationPercent = 100.0 / totalDuration * current.Duration
		current.SetID()
		out = append(out, current)
	}

	// get last state before the time window specified by 'start' and 'end'
	gap, _ := time.ParseDuration("30m")
	q = newSelectQuery().From(rs.Kind()).Between(end.Add(gap*-1), end).FilterTags(rs.Tags).OrderBy("time").Desc().Limit(1)
	last, err := i.First(q)

	if err != nil {
		// if no state at all is found
		complete := store.Span{
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
		out = append([]store.Span{complete}, out...)
	} else {
		duration := earliestSpan.Sub(start).Seconds()
		durationPercent := 100.0 / totalDuration * duration
		first := store.Span{
			Start:           start,
			End:             earliestSpan,
			Pseudo:          true,
			Duration:        duration,
			DurationPercent: durationPercent,
			Tags:            last.Tags,
			Status:          last.Status,
			Annotation:      last.Annotation,
		}
		first.SetID()
		out = append([]store.Span{first}, out...)
	}

	return out, nil
}

func (i Influx) assumeEvents(rs store.ResultSet, start time.Time, end time.Time, interval time.Duration) ([]store.Span, error) {
	duration := end.Sub(start).Seconds()
	s := []store.Span{
		store.Span{
			Status:     status.OK,
			Annotation: "",
			Start:      start,
			End:        end,
			Tags:       rs.Tags,
		},
	}

	q := newSelectQuery().From(rs.Kind()).Between(start, end).FilterTags(rs.Tags)
	rows, err := i.Run(q)
	if err != nil {
		msg := fmt.Sprintf("Cannot run query, error is: %s", err.Error())
		return s, errors.New(msg)
	}

	lastIndex := len(s) - 1

	for _, row := range rows {
		last := s[lastIndex]
		replace := false

		current := store.Span{
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
				s[lastIndex] = last
			}
		} else if current.Start.After(last.End) {
			filler := store.Span{
				Start:      last.End,
				End:        current.Start,
				Status:     status.OK,
				Annotation: "",
			}
			s = append(s, filler)
		}

		// igrnore that case for now
		// if current.End.Before(last.End) {}

		if replace {
			s[lastIndex] = current
		} else {
			s = append(s, current)
		}

		lastIndex = len(s) - 1
	}

	lastEvent := s[len(s)-1]
	if lastEvent.End != end {
		filler := store.Span{
			Start:      lastEvent.End,
			End:        end,
			Status:     status.OK,
			Annotation: "",
		}
		s = append(s, filler)
	}

	for i, e := range s {
		e.Duration = e.End.Sub(e.Start).Seconds()
		e.DurationPercent = 100.0 / duration * e.Duration
		e.SetID()
		s[i] = e
	}
	return s, nil
}