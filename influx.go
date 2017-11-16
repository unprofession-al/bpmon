package bpmon

import (
	"errors"
	"fmt"
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

type Point struct {
	Timestamp time.Time
	Series    string
	Tags      map[string]string
	Fields    map[string]interface{}
}
