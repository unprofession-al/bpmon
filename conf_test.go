package bpmon

import (
	"testing"
	"time"
)

var confyaml = []byte(`---
default: &default
  icinga:
    server: icinca.example.com
    port: 5665
    pass: pa55w0rd
    user: bpmon
    proto: https
  influx:
    connection:
      server: influx.example.com
      port: 8086
      proto: http
    database: bpmon
    save_ok: [ BP ]
  availabilities:
    7x24:
      monday:    [ "allday" ]
      tuesday:   [ "allday" ]
      wednesday: [ "allday" ]
      thursday:  [ "allday" ]
      friday:    [ "allday" ]
      saturday:  [ "allday" ]
      sunday:    [ "allday" ]
test:
  <<: *default
  influx:
    database: bpmon_test
  availabilities:
    9to5:
      monday:    [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      tuesday:   [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      wednesday: [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      thursday:  [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      friday:    [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
`)

var bpconfig = []byte(`---
name: Test BP
id: testbp
availability: Test
kpis:
  - name: Test KPI
    id: testkpi
    operation: AND
    services:
      - host: test.example.com
        service: svc0
      - host: test.example.com
        service: svc1
      - host: test.example.com
        service: svc2
`)

func TestParseConf(t *testing.T) {
	c, err := parseConf(confyaml, "test")
	if err != nil {
		t.Errorf("Could not parse config: %s", err.Error())
	}
	if _, ok := c.Availabilities["9to5"]; !ok {
		t.Errorf("Config not parsed correctly")
	}
}

func TestParseConfUnknownSection(t *testing.T) {
	_, err := parseConf(confyaml, "foo")
	if err == nil {
		t.Errorf("Found section 'foo' that does not exist")
	}
}

func TestParseBP(t *testing.T) {
	a := Availabilities{
		"Test": Availability{
			time.Monday: AvailabilityTime{
				AllDay: true,
			},
			time.Friday: AvailabilityTime{
				AllDay: true,
			},
		},
	}
	_, err := parseBP(bpconfig, a)
	if err != nil {
		t.Errorf("Could not parse BP config: %s", err.Error())
	}
}

func TestParseBPUnknownAvailability(t *testing.T) {
	a := Availabilities{
		"Never": Availability{},
	}
	_, err := parseBP(bpconfig, a)
	if err == nil {
		t.Errorf("Found availability 'Test' that does not exist")
	}
}
