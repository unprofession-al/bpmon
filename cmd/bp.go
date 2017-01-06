package cmd

import (
	"fmt"
	"strings"
	"time"
)

type bp struct {
	Name string
	Id   string
	Kpis []kpi
}

type ServiceStatusProvider interface {
	ServiceStatus(Service) (bool, error)
}

func (bp bp) Status(ssp ServiceStatusProvider) ResultSet {
	rs := ResultSet{
		kind:     "BP",
		name:     bp.Name,
		id:       bp.Id,
		children: []ResultSet{},
	}

	ch := make(chan *ResultSet)
	var values []bool
	for _, k := range bp.Kpis {
		go func(k kpi, ssp ServiceStatusProvider) {
			childRs := k.Status(ssp)
			ch <- &childRs
		}(k, ssp)
	}

	for {
		select {
		case childRs := <-ch:
			values = append(values, childRs.Bool())
			rs.children = append(rs.children, *childRs)
			if len(values) == len(bp.Kpis) {
				ch = nil
			}
		}
		if ch == nil {
			break
		}
	}

	ok, err := calculate("AND", values)
	rs.status = boolAsStatus(ok)
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
	Services  []Service
}

func (k kpi) Status(ssp ServiceStatusProvider) ResultSet {
	rs := ResultSet{
		kind:     "KPI",
		name:     k.Name,
		id:       k.Id,
		children: []ResultSet{},
	}

	ch := make(chan *ResultSet)
	var values []bool
	for _, s := range k.Services {
		go func(s Service, ssp ServiceStatusProvider) {
			childRs := s.Status(ssp)
			ch <- &childRs
		}(s, ssp)
	}

	for {
		select {
		case childRs := <-ch:
			values = append(values, childRs.Bool())
			rs.children = append(rs.children, *childRs)
			if len(values) == len(k.Services) {
				ch = nil
			}
		}
		if ch == nil {
			break
		}
	}

	ok, err := calculate(k.Operation, values)
	rs.status = boolAsStatus(ok)
	if err != nil {
		rs.err = err
		rs.status = StatusUnknown
	}
	return rs
}

type Service struct {
	Host    string
	Service string
}

func (s Service) Status(ssp ServiceStatusProvider) ResultSet {
	name := fmt.Sprintf("%s!%s", s.Host, s.Service)
	rs := ResultSet{
		name: name,
		id:   name,
		kind: "SVC",
	}
	ok, err := ssp.ServiceStatus(s)
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
	id       string
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

func (rs ResultSet) AsInflux(nt map[string]string, t time.Time) []Point {
	var out []Point

	nt[rs.kind] = rs.id
	tags := map[string]string{
		"kind": rs.kind,
	}
	for k, v := range nt {
		tags[k] = v
	}
	fields := map[string]interface{}{
		"status": rs.status.toInt(),
	}
	pt := Point{
		Series: rs.kind,
		Tags:   tags,
		Fields: fields,
		Time:   t,
	}
	out = append(out, pt)

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
