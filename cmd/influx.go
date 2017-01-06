package cmd

import (
	"fmt"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxConf struct {
	Connection struct {
		Server string
		Port   int
		Pass   string
		User   string
		Proto  string
	}
	SaveOk bool
}

type Influx struct {
	cli    client.Client
	saveOk bool
}

func NewInflux(conf InfluxConf) (Influx, error) {
	addr := fmt.Sprintf("%s://%s:%d", conf.Connection.Proto, conf.Connection.Server, conf.Connection.Port)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: conf.Connection.User,
		Password: conf.Connection.Pass,
	})
	cli := Influx{
		cli:    c,
		saveOk: conf.SaveOk,
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
