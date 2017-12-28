package influx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/unprofession-al/bpmon/store"
)

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
	out := store.ResultSet{}
	query := NewSelectQuery().From(rs.Kind()).FilterTags(rs.Tags).OrderBy("time").Desc().Limit(1)
	data, err := i.First(query)
	out, err = i.asResultSet(data)
	return out, err
}

func (i Influx) First(query Query) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	all, err := i.Run(query)
	if err != nil {
		return out, err
	}

	if len(all) == 0 {
		return out, errors.New("no data returned")
	}

	for k, v := range all[0] {
		out[k] = v
	}

	return out, nil
}

func (i Influx) Run(query Query) ([]map[string]interface{}, error) {
	var out []map[string]interface{}

	if i.printQueries {
		fmt.Println(query.Query())
	}

	q := client.Query{
		Command:  query.Query(),
		Database: i.database,
	}

	response, err := i.cli.Query(q)
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
			out = append(out, data)
		}

	}
	return out, nil
}
