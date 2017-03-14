package bpmon

import "testing"

var conf_yaml = []byte(`
---
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
	9to5:
      monday:    [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      tuesday:   [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      wednesday: [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      thursday:  [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]
      friday:    [ "09:00:00-12:00:00", "13:00:00-17:00:00" ]

test:
  <<: *default
  influx:
    database: bpmon_test
`)

func TestReadConf(t *testing.T) {
	t.Log("Not implemented")
}
