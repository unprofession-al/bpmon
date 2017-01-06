package cmd

import (
	"fmt"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxConf struct {
	Server string
	Port   int
	Pass   string
	User   string
	Proto  string
}

type Influx struct {
	cli client.Client
}

func NewInflux(conf InfluxConf) (Influx, error) {
	addr := fmt.Sprintf("%s://%s:%d", conf.Proto, conf.Server, conf.Port)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: conf.User,
		Password: conf.Pass,
	})
	cli := Influx{
		cli: c,
	}
	return cli, err
}

func (i Influx) Write(rs ResultSet, ts time.Time) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "bpmon",
		Precision: "s",
	})
	if err != nil {
		return err
	}

	ns := make(map[string]string)
	points := rs.AsInflux(ns, ts)

	for _, p := range points {
		pt, _ := client.NewPoint(p.Series, p.Tags, p.Fields, p.Time)
		bp.AddPoint(pt)
	}
	err = i.cli.Write(bp)

	return err
}

type Point struct {
	Series string
	Tags   map[string]string
	Fields map[string]interface{}
	Time   time.Time
}
