package cmd

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type Influx struct {
	cli client.Client
}

func NewInflux(conf client.HTTPConfig) (Influx, error) {
	c, err := client.NewHTTPClient(conf)
	cli := Influx{
		cli: c,
	}
	return cli, err
}

func (i Influx) writeResultSet(rs ResultSet) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "bpmon",
		Precision: "s",
	})
	if err != nil {
		return err
	}

	ns := make(map[string]string)
	points := rs.AsInflux(ns, time.Now())

	for _, p := range points {
		pt, _ := client.NewPoint(p.series, p.tags, p.fields, p.time)
		bp.AddPoint(pt)
	}
	err = i.cli.Write(bp)

	return err
}

type point struct {
	series string
	tags   map[string]string
	fields map[string]interface{}
	time   time.Time
}
