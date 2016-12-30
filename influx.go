package main

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
		Precision: "m",
	})
	if err != nil {
		return err
	}

	ns := make(map[string]string)
	points := rs.AsInflux(ns, time.Now())

	for _, pt := range points {
		bp.AddPoint(&pt)
	}
	err = i.cli.Write(bp)

	return err
}
