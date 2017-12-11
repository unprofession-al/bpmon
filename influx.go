package bpmon

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/unprofession-al/bpmon/status"
)

type InfluxConf struct {
	Connection struct {
		Server  string        `yaml:"server"`
		Port    int           `yaml:"port"`
		Pass    string        `yaml:"pass"`
		User    string        `yaml:"user"`
		Proto   string        `yaml:"proto"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"connection"`
	SaveOK        []string               `yaml:"save_ok"`
	Database      string                 `yaml:"database"`
	GetLastStatus bool                   `yaml:"get_last_status"`
	DefaultTags   map[string]string      `yaml:"default_tags"`
	DefaultFields map[string]interface{} `yaml:"default_fields"`
}

type Influx struct {
	cli           client.Client
	saveOK        []string
	database      string
	defaultTags   map[string]string
	defaultFields map[string]interface{}
}

type Influxable interface {
	AsInflux([]string, map[string]string, map[string]interface{}) []Point
}

type Point struct {
	Timestamp time.Time              `json:"timestamp"`
	Series    string                 `json:"series"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

func NewInflux(conf InfluxConf) (Influx, error) {
	addr := fmt.Sprintf("%s://%s:%d", conf.Connection.Proto, conf.Connection.Server, conf.Connection.Port)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: conf.Connection.User,
		Password: conf.Connection.Pass,
		Timeout:  conf.Connection.Timeout,
	})
	cli := Influx{
		cli:           c,
		saveOK:        conf.SaveOK,
		defaultTags:   conf.DefaultTags,
		defaultFields: conf.DefaultFields,
		database:      conf.Database,
	}
	return cli, err
}

func (i Influx) Write(in Influxable, debug bool) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	points := in.AsInflux(i.saveOK, i.defaultTags, i.defaultFields)

	for _, p := range points {
		pt, _ := client.NewPoint(p.Series, p.Tags, p.Fields, p.Timestamp)
		bp.AddPoint(pt)
	}
	if debug {
		for _, p := range bp.Points() {
			fmt.Println(p)
		}
	} else {
		err = i.cli.Write(bp)
	}

	return err
}

func (i Influx) GetOne(query string) (interface{}, error) {
	var out interface{}
	q := client.Query{
		Command:  query,
		Database: i.database,
	}

	response, err := i.cli.Query(q)
	if err != nil {
		return out, err
	}
	if response.Error() != nil {
		return out, response.Error()
	}

	if len(response.Results) >= 1 &&
		len(response.Results[0].Series) >= 1 &&
		len(response.Results[0].Series[0].Values) >= 1 &&
		len(response.Results[0].Series[0].Values[0]) >= 2 {
		out = response.Results[0].Series[0].Values[0][1]
	} else {
		err = errors.New("No earlier entry found")
		return out, err
	}

	return out, nil
}

func getKind(spec map[string]string) string {
	kind := "UNKNOWN"
	if _, ok := spec[IdentifierBusinessProcess]; ok {
		kind = IdentifierBusinessProcess
	}
	if _, ok := spec[IdentifierKeyPerformanceIndicator]; ok {
		kind = IdentifierKeyPerformanceIndicator
	}
	if _, ok := spec[IdentifierService]; ok {
		kind = IdentifierService
	}
	return kind
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

type ev struct {
	Status          status.Status `json:"status"`
	Annotation      string        `json:"annotation"`
	Start           time.Time     `json:"start"`
	End             time.Time     `json:"end"`
	Duration        float64       `json:"duration"`
	DurationPercent float64       `json:"duration_percent"`
}

func (i Influx) AssumeEvents(spec map[string]string, start time.Time, end time.Time, interval time.Duration) ([]Point, error) {
	kind := getKind(spec)
	startTs := getInfluxTimestamp(start)
	endTs := getInfluxTimestamp(end)
	duration := end.Sub(start).Seconds()
	events := []ev{
		ev{Status: status.Ok,
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

			current := ev{}
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
				filler := ev{
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
		filler := ev{
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

func getInfluxTimestamp(t time.Time) int64 {
	return t.UnixNano()
}

func queryDB(clnt client.Client, db string, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: db,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
