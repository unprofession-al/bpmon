package bpmon

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestStringToWeekday(t *testing.T) {
	tests := []struct {
		str         string
		day         time.Weekday
		errExpected bool
	}{
		{str: "monday", day: time.Monday, errExpected: false},
		{str: "Tuesday", day: time.Tuesday, errExpected: false},
		{str: "Wednesday", day: time.Wednesday, errExpected: false},
		{str: "Thursday", day: time.Thursday, errExpected: false},
		{str: "FRIDAY", day: time.Friday, errExpected: false},
		{str: "Saturday", day: time.Saturday, errExpected: false},
		{str: "Sunday", day: time.Sunday, errExpected: false},
		{str: "Casual-Friday", errExpected: true},
	}

	for _, test := range tests {
		day, err := toWeekday(test.str)
		if err == nil && test.errExpected == true {
			t.Errorf("Error expected for '%s' but test succeeded", test.str)
		} else if err != nil && test.errExpected == false {
			t.Errorf("No error expected for '%s' but test failed: %s", test.str, err.Error())
		} else if err == nil && test.errExpected == false {
			if day != test.day {
				t.Errorf("Result not as expected for '%s': Should be '%v', is '%v'", test.str, test.day, day)
			}
		}
	}
}

func ParseTime(str string) time.Time {
	format := "15:04:05.000"
	t, err := time.Parse(format, str)
	if err != nil {
		panic(fmt.Sprintf("Time in test malformed, is '%s', must match '%s', error is: %s", str, format, err.Error()))
	}
	return t
}

func TestStringsToAvailabilityTime(t *testing.T) {
	tests := []struct {
		str         []string
		at          AvailabilityTime
		errExpected bool
	}{
		{
			str: []string{"09:00:00-12:00:00"},
			at: AvailabilityTime{
				TimeRanges: []TimeRange{
					{Start: ParseTime("09:00:00.000"), End: ParseTime("12:00:00.000")},
				},
				AllDay: false,
			},
			errExpected: false,
		},
		{
			str:         []string{},
			errExpected: true,
		},
		{
			str:         []string{"12:00:00"},
			errExpected: true,
		},
		//{
		//	str: []string{"ALLDAY", "09:00:00-12:00:00"},
		//	at: AvailabilityTime{
		//		AllDay: true,
		//	},
		//	errExpected: false,
		//},
	}

	for _, test := range tests {
		at, err := toAvailabilityTime(test.str)
		if err == nil && test.errExpected == true {
			t.Errorf("Error expected for '%s' but test succeeded", test.str)
		} else if err != nil && test.errExpected == false {
			t.Errorf("No error expected for '%s' but test failed: %s", test.str, err.Error())
		} else if err == nil && test.errExpected == false {
			eq := reflect.DeepEqual(at, test.at)
			if !eq {
				t.Errorf("Results do not match for %v: '%v' vs. '%v'", test.str, at, test.at)
			}
		}
	}
}
