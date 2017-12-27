package influx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/unprofession-al/bpmon/persistence"
)

type Influx struct {
	cli           client.Client
	saveOK        []string
	database      string
	printQueries  bool
	getLastStatus bool
}

func init() {
	persistence.Register("influx", Setup)
}

func Setup(conf persistence.Conf) (persistence.Persistence, error) {
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

func (i Influx) Write(rs *persistence.ResultSet) error {
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

func (i Influx) GetLatest(rs persistence.ResultSet) (persistence.ResultSet, error) {
	out := persistence.ResultSet{}
	data := make(map[string]interface{})

	var where []string
	for key, value := range rs.Tags {
		where = append(where, fmt.Sprintf("%s = '%s'", key, value))
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY time DESC LIMIT 1;", rs.Kind(), strings.Join(where, " AND "))

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

	var fields []string
	if len(response.Results) >= 1 && len(response.Results[0].Series) >= 1 {
		fields = response.Results[0].Series[0].Columns
	}

	if len(response.Results) >= 1 &&
		len(response.Results[0].Series) >= 1 &&
		len(response.Results[0].Series[0].Values) >= 1 &&
		len(response.Results[0].Series[0].Values[0]) >= 2 {
		row := response.Results[0].Series[0].Values[0]
		for i, field := range row {
			data[fields[i]] = field
		}
	} else {
		err = errors.New("No matching entry found")
		return out, err
	}

	out, err = i.asResultSet(data)
	return out, err
}

func (i Influx) getOne(fields []string, from string, where []string, additional string) (map[string]interface{}, error) {
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

func (i Influx) getAll(fields []string, from string, where []string, additional string) ([]map[string]interface{}, error) {
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
