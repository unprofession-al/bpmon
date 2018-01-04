package influx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/unprofession-al/bpmon/store"
)

const timefield = "time"

type Influx struct {
	cli           client.Client
	saveOK        []string
	database      string
	printQueries  bool
	getLastStatus bool
}

func init() {
	store.Register("influx", Setup)
}

func Setup(conf store.Conf) (store.Store, error) {
	u, err := url.Parse(conf.Connection)
	if err != nil {
		panic(err)
	}
	database := strings.TrimLeft(u.Path, "/")
	username := u.User.Username()
	password, _ := u.User.Password()

	addr := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
		Timeout:  conf.Timeout,
	})
	cli := Influx{
		cli:           c,
		saveOK:        conf.SaveOK,
		database:      database,
		printQueries:  conf.Debug,
		getLastStatus: conf.GetLastStatus,
	}

	return cli, err
}

func (i Influx) Write(rs *store.ResultSet) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	points := i.asPoints(rs)
	for _, p := range points {
		pt, _ := client.NewPoint(p.Series, p.Tags, p.Fields, p.Timestamp)
		bp.AddPoint(pt)
	}

	if i.printQueries {
		for _, p := range bp.Points() {
			fmt.Println(p)
		}
	} else {
		err = i.cli.Write(bp)
	}

	return err
}

func (i Influx) GetLatest(rs store.ResultSet) (store.ResultSet, error) {
	q := newSelectQuery().From(rs.Kind().String()).FilterTags(rs.Tags).OrderBy("time").Desc().Limit(1)
	return i.First(q)
}

func (i Influx) First(q query) (store.ResultSet, error) {
	var out store.ResultSet

	all, err := i.Run(q)
	if err != nil {
		return out, err
	}

	if len(all) == 0 {
		return out, errors.New("no data returned")
	}

	out = all[0]
	return out, nil
}

func (i Influx) Run(q query) ([]store.ResultSet, error) {
	var out []store.ResultSet

	if i.printQueries {
		fmt.Println(q.Query())
	}

	cq := client.Query{
		Command:  q.Query(),
		Database: i.database,
	}

	response, err := i.cli.Query(cq)
	if err != nil {
		return out, err
	}
	if response.Error() != nil {
		return out, response.Error()
	}

	var fields []string
	if len(response.Results) >= 1 && len(response.Results[0].Series) >= 1 {
		fields = response.Results[0].Series[0].Columns
	}

	if len(response.Results) >= 1 &&
		len(response.Results[0].Series) >= 1 &&
		len(response.Results[0].Series[0].Values) >= 1 {
		rows := response.Results[0].Series[0].Values
		for _, row := range rows {
			data := make(map[string]interface{})
			for i, cell := range row {
				data[fields[i]] = cell
			}
			rs, err := i.asResultSet(data)
			if err != nil {
				return out, err
			}
			out = append(out, rs)
		}

	}
	return out, nil
}
