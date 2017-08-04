package bpmon

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/mgutz/ansi"
	"github.com/unprofession-al/bpmon/status"
)

type ResultSet struct {
	Name          string
	Id            string
	Kind          string
	At            time.Time
	Vals          map[string]bool
	Status        status.Status
	Was           status.Status
	WasChecked    bool
	StatusChanged bool
	Err           error
	Output        string
	Children      []*ResultSet
}

func (rs ResultSet) PrettyPrint(level int, ts bool, vals bool) string {
	ident := strings.Repeat("   ", level)
	out := rs.Status.Colorize(fmt.Sprintf("%s%s %s is %v", ident, rs.Kind, rs.Name, rs.Status))
	if rs.WasChecked {
		out += rs.Status.Colorize(fmt.Sprintf(" (was %v)", rs.Was))
	}

	ident = strings.Repeat("   ", level+1)
	if ts {
		timestamp := rs.At.Format("2006-01-02 15:04:05")
		out += fmt.Sprintf(" (%s)", timestamp)
	}
	if rs.Err != nil {
		out += fmt.Sprintf("\n%sError occured: %s", ident, rs.Err.Error())
	}
	if rs.Status == status.Nok && rs.Output != "" {
		out += fmt.Sprintf("\n%sMessage from Monitoring: %s", ident, rs.Output)
	}
	if vals && len(rs.Vals) > 0 {
		out += fmt.Sprintf("\n%sValues:", ident)
		var keys []string
		for k := range rs.Vals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := rs.Vals[key]
			out += " "
			if !value {
				out += ansi.Color(key, "magenta+s")
			} else {
				out += ansi.Color(key, "green")
			}
		}
	}
	out += "\n"
	for _, childRs := range rs.Children {
		out += childRs.PrettyPrint(level+1, ts, vals)
	}
	return out
}

func (rs ResultSet) StripByStatus(s []status.Status) (ResultSet, bool) {
	setOut := rs
	keep := true
	for _, status := range s {
		if rs.Status == status {
			keep = false
			break
		}
	}
	if keep {
		var children []*ResultSet
		for _, child := range rs.Children {
			set, stripped := child.StripByStatus(s)
			if !stripped {
				children = append(children, &set)
			}
		}
		setOut.Children = children
	}
	return setOut, !keep
}

func (rs ResultSet) AsInflux(saveOK []string) []Point {
	tags := make(map[string]string)
	return rs.toPoints(tags, saveOK)
}

func (rs ResultSet) toPoints(parentTags map[string]string, saveOK []string) []Point {
	var out []Point

	tags := make(map[string]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[rs.Kind] = rs.Id

	if rs.Status != status.Ok || stringInSlice(rs.Kind, saveOK) {
		fields := map[string]interface{}{
			"status": rs.Status.Int(),
		}
		for key, value := range rs.Vals {
			fields[key] = value
		}
		if rs.Output != "" {
			fields["output"] = fmt.Sprintf("Output: %s", rs.Output)
		}
		if rs.Err != nil {
			fields["err"] = fmt.Sprintf("Error: %s", rs.Err.Error())
		}
		if rs.WasChecked {
			fields["was"] = rs.Was.Int()
			fields["changed"] = rs.StatusChanged
		}
		pt := Point{
			Timestamp: rs.At,
			Series:    rs.Kind,
			Tags:      tags,
			Fields:    fields,
		}
		out = append(out, pt)
	}

	for _, childRs := range rs.Children {
		out = append(out, childRs.toPoints(tags, saveOK)...)
	}
	return out
}

func (rs *ResultSet) AddPreviousStatus(pp PersistenceProvider, saveOK []string) {
	tags := make(map[string]string)
	rs.previousStatus(tags, pp, saveOK)
}

func (rs *ResultSet) previousStatus(parentTags map[string]string, pp PersistenceProvider, saveOK []string) {
	tags := make(map[string]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[rs.Kind] = rs.Id

	if stringInSlice(rs.Kind, saveOK) {
		was, err := getLastStatus(pp, rs.Kind, tags)
		if err == nil {
			rs.Was = was
			rs.WasChecked = true
			if rs.Status != rs.Was {
				rs.StatusChanged = true
			}
		}
	}

	for _, childRs := range rs.Children {
		childRs.previousStatus(tags, pp, saveOK)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToUpper(b) == strings.ToUpper(a) {
			return true
		}
	}
	return false
}

func getLastStatus(pp PersistenceProvider, kind string, tags map[string]string) (status.Status, error) {
	var wheres []string
	for key, value := range tags {
		wheres = append(wheres, fmt.Sprintf("%s = '%s'", key, value))
	}
	where := strings.Join(wheres, " AND ")
	query := fmt.Sprintf("SELECT LAST(status) FROM %s WHERE %s", kind, where)
	res, err := pp.GetOne(query)
	if err != nil {
		return status.Unknown, err
	}

	statusData, ok := res.(json.Number)
	if !ok {
		msg := fmt.Sprintf("Cannot convert %v (%s) to json.Number", res, reflect.TypeOf(res))
		return status.Unknown, errors.New(msg)
	}

	statusCode, err := statusData.Int64()
	if err != nil {
		return status.Unknown, err
	}

	return status.FromInt64(statusCode)
}
