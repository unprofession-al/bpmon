package bpmon

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
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
	SaveOK        []string `yaml:"save_ok"`
	Database      string   `yaml:"database"`
	GetLastStatus bool     `yaml:"get_last_status"`
	PrintQueries  bool     `yaml:"print_queries"`
}

type Influx struct {
	cli          client.Client
	saveOK       []string
	database     string
	printQueries bool
}

type Influxable interface {
	AsInflux([]string) []Point
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
		cli:          c,
		saveOK:       conf.SaveOK,
		database:     conf.Database,
		printQueries: conf.PrintQueries,
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

	points := in.AsInflux(i.saveOK)

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
func (i Influx) WritePoints(points []Point, debug bool) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

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

func (i Influx) GetOne(fields []string, from string, where []string, additional string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	fields = prependTimeIfMissing(fields)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s %s;", strings.Join(fields, ", "), from, strings.Join(where, " AND "), additional)
	if i.printQueries {
		fmt.Println(query)
	}

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

	if len(fields) == 1 && fields[0] == "*" {
		if len(response.Results) >= 1 && len(response.Results[0].Series) >= 1 {
			fields = response.Results[0].Series[0].Columns
		}
	}

	if len(response.Results) >= 1 &&
		len(response.Results[0].Series) >= 1 &&
		len(response.Results[0].Series[0].Values) >= 1 &&
		len(response.Results[0].Series[0].Values[0]) >= 2 {
		row := response.Results[0].Series[0].Values[0]
		for i, data := range row {
			out[fields[i]] = data
		}
	} else {
		err = errors.New("No matching entry found")
		return out, err
	}

	return out, nil
}

func (i Influx) GetAll(fields []string, from string, where []string, additional string) ([]map[string]interface{}, error) {
	var out []map[string]interface{}

	fields = prependTimeIfMissing(fields)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s %s;", strings.Join(fields, ", "), from, strings.Join(where, " AND "), additional)
	if i.printQueries {
		fmt.Println(query)
	}

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

	if len(fields) == 1 && fields[0] == "*" {
		if len(response.Results) >= 1 && len(response.Results[0].Series) >= 1 {
			fields = response.Results[0].Series[0].Columns
		}
	}

	if len(response.Results) >= 1 &&
		len(response.Results[0].Series) >= 1 &&
		len(response.Results[0].Series[0].Values) >= 1 {
		rows := response.Results[0].Series[0].Values
		for _, row := range rows {
			set := make(map[string]interface{})
			for i, data := range row {
				set[fields[i]] = data
			}
			out = append(out, set)
		}

	}

	return out, nil
}

func prependTimeIfMissing(fields []string) []string {
	for i, field := range fields {
		if field == "*" {
			return []string{"*"}
		}
		if field == "time" {
			if i != 0 {
				tmp := fields[0]
				fields[0] = "time"
				fields[i] = tmp
			}
			return fields
		}
	}
	return append([]string{"time"}, fields...)
}

func getInfluxTimestamp(t time.Time) int64 {
	return t.UnixNano()
}
