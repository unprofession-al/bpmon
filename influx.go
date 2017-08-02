package bpmon

import (
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/unprofession-al/bpmon/status"
)

type InfluxConf struct {
	Connection struct {
		Server string `yaml:"server"`
		Port   int    `yaml:"port"`
		Pass   string `yaml:"pass"`
		User   string `yaml:"user"`
		Proto  string `yaml:"proto"`
	} `yaml:"connection"`
	SaveOK   []string `yaml:"save_ok"`
	Database string   `yaml:"database"`
}

type Influx struct {
	cli      client.Client
	saveOK   []string
	database string
}

type Influxable interface {
	AsInflux([]string) []Point
}

func NewInflux(conf InfluxConf) (Influx, error) {
	addr := fmt.Sprintf("%s://%s:%d", conf.Connection.Proto, conf.Connection.Server, conf.Connection.Port)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: conf.Connection.User,
		Password: conf.Connection.Pass,
	})
	cli := Influx{
		cli:      c,
		saveOK:   conf.SaveOK,
		database: conf.Database,
	}
	return cli, err
}

func (i Influx) Write(in Influxable) error {
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
	err = i.cli.Write(bp)

	return err
}

func (i Influx) GetLastStatus(bp string) (status.Status, error) {
	var out status.Status
	out = status.Unknown
	q := client.Query{
		Command:  fmt.Sprintf("SELECT LAST(status) FROM BP WHERE BP = '%s'", bp),
		Database: i.database,
	}

	response, err := i.cli.Query(q)
	if err != nil {
		return out, err
	}
	if response.Error() != nil {
		return out, response.Error()
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("No earlier entry found")
			return
		}
	}()

	bla := response.Results[0].Series[0].Values[0][1]
	out = bla.(status.Status)
	return out, nil
}

type Point struct {
	Timestamp time.Time
	Series    string
	Tags      map[string]string
	Fields    map[string]interface{}
}
