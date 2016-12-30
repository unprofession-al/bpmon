package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type bp struct {
	Name string
	Id   string
	Kpis []kpi
}

func (bp bp) Status() ResultSet {
	rs := ResultSet{
		kind:     "BP",
		name:     bp.Name,
		children: []ResultSet{},
	}

	var values []bool
	for _, kpi := range bp.Kpis {
		childRs := kpi.Status()
		values = append(values, childRs.Bool())
		rs.children = append(rs.children, childRs)
	}

	ok, err := calculate("AND", values)
	rs.status = boolToStatus(ok)
	if err != nil {
		rs.err = err
		rs.status = StatusUnknown
	}
	return rs
}

type kpi struct {
	Name      string
	Id        string
	Operation string
	Services  []service
}

func (k kpi) Status() ResultSet {
	rs := ResultSet{
		kind:     "KPI",
		name:     k.Name,
		children: []ResultSet{},
	}

	var values []bool
	for _, s := range k.Services {
		childRs := s.Status()
		values = append(values, childRs.Bool())
		rs.children = append(rs.children, childRs)
	}

	ok, err := calculate(k.Operation, values)
	rs.status = boolToStatus(ok)
	if err != nil {
		rs.err = err
		rs.status = StatusUnknown
	}
	return rs
}

type service struct {
	Host    string
	Service string
}

func (s service) Status() ResultSet {
	name := fmt.Sprintf("%s!%s", s.Host, s.Service)
	rs := ResultSet{
		name: name,
		kind: "SVC",
	}
	ok, err := i.ServiceStatus(s.Host, s.Service)
	rs.err = err
	if rs.err != nil {
		rs.status = StatusUnknown
	} else if ok {
		rs.status = StatusOK
	} else {
		rs.status = StatusNOK
	}
	return rs
}

type ResultSet struct {
	name     string
	kind     string
	status   status
	err      error
	children []ResultSet
}

func (rs ResultSet) PrettyPrint(level int) string {
	ident := strings.Repeat("   ", level)
	out := rs.status.Colorize(fmt.Sprintf("%s%s %s is %v", ident, rs.kind, rs.name, rs.status))
	if rs.err != nil {
		out += fmt.Sprintf(" (Error occured: %s)", rs.err.Error())
	}
	out += "\n"
	for _, childRs := range rs.children {
		out += childRs.PrettyPrint(level + 1)
	}
	return out
}

type nameTags map[string]string

func (rs ResultSet) AsInflux(nt nameTags, t time.Time) []client.Point {
	var out []client.Point

	nt[rs.kind] = rs.name
	tags := map[string]string{
		"kind": rs.kind,
	}
	for k, v := range nt {
		tags[k] = v
	}
	fields := map[string]interface{}{
		"status": rs.status,
	}
	pt, _ := client.NewPoint(nt["BP"], tags, fields, t)
	out = append(out, *pt)

	for _, childRs := range rs.children {
		out = append(out, childRs.AsInflux(nt, t)...)
	}
	return out
}

func (rs ResultSet) Bool() bool {
	if rs.status == StatusNOK {
		return false
	}
	return true
}
