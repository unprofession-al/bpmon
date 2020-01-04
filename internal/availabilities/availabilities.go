package availabilities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const timeformat = "15:04:05"

type AvailabilitiesConfig map[string]AvailabilityConfig

func (ac AvailabilitiesConfig) Parse() (Availabilities, error) {
	a := make(Availabilities)
	for name, daysConf := range ac {
		availability, err := daysConf.Parse()
		if err != nil {
			return a, err
		}
		a[name] = availability
	}
	return a, nil
}

type Availabilities map[string]Availability

type AvailabilityConfig map[string][]string

func (ac AvailabilityConfig) Parse() (Availability, error) {
	a := make(Availability)
	for day, tsString := range ac {
		wd, err := toWeekday(day)
		if err != nil {
			return a, err
		}

		at, err := toAvailabilityTime(tsString)
		if err != nil {
			return a, err
		}

		a[wd] = at
	}
	return a, nil
}

func toWeekday(str string) (time.Weekday, error) {
	switch strings.Title(strings.ToLower(str)) {
	case time.Sunday.String():
		return time.Sunday, nil
	case time.Monday.String():
		return time.Monday, nil
	case time.Tuesday.String():
		return time.Tuesday, nil
	case time.Wednesday.String():
		return time.Wednesday, nil
	case time.Thursday.String():
		return time.Thursday, nil
	case time.Friday.String():
		return time.Friday, nil
	case time.Saturday.String():
		return time.Saturday, nil
	default:
		return time.Monday, fmt.Errorf("'%s' does not look like the name of a weekday", str)
	}
}

type Availability map[time.Weekday]AvailabilityTime

func (a Availability) Contains(t time.Time) bool {
	wd := t.Weekday()
	at, ok := a[wd]

	if !ok {
		return false
	}

	if at.AllDay {
		return true
	}

	for _, tRange := range a[wd].TimeRanges {
		year, month, day := t.Date()
		startH, startM, startS := tRange.Start.Clock()
		endH, endM, endS := tRange.End.Clock()
		loc := t.Location()

		start := time.Date(year, month, day, startH, startM, startS, 0, loc)
		end := time.Date(year, month, day, endH, endM, endS, 0, loc)

		if t.After(start) && t.Before(end) {
			return true
		}
	}
	return false
}

type AvailabilityTime struct {
	TimeRanges []TimeRange
	AllDay     bool
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func toAvailabilityTime(trStrings []string) (AvailabilityTime, error) {
	out := AvailabilityTime{AllDay: false}
	if len(trStrings) < 1 || trStrings[0] == "" {
		return out, errors.New("time range definition cannot be empty")
	}
	timeranges := make([]TimeRange, 0)
	for _, trString := range trStrings {
		tStrings := strings.Split(trString, "-")
		if strings.ToLower(tStrings[0]) == "allday" {
			out.AllDay = true
			break
		}
		if len(tStrings) != 2 {
			return out, fmt.Errorf("'%s' does not look like a time range definition, time slots must be formated as in '%s-%s'", trString, timeformat, timeformat)
		}
		start, err := toTime(tStrings[0])
		if err != nil {
			return out, err
		}
		end, err := toTime(tStrings[1])
		if err != nil {
			return out, err
		}
		timeranges = append(timeranges, TimeRange{
			Start: start,
			End:   end,
		})
	}
	out.TimeRanges = timeranges
	return out, nil
}

func toTime(tString string) (time.Time, error) {
	tString = strings.TrimSpace(tString)
	t, err := time.Parse(timeformat, tString)
	if err != nil {
		return t, fmt.Errorf("'%s' does not look like a time, times must be formated as in '%s'", tString, timeformat)
	}
	return t, err

}
