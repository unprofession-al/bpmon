package bpmon

import (
	"fmt"
	"time"

	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
)

type BusinessProcesses []BP

type BP struct {
	Name             string       `yaml:"name"`
	Id               string       `yaml:"id"`
	Kpis             []KPI        `yaml:"kpis"`
	AvailabilityName string       `yaml:"availability"`
	Availability     Availability `yaml:"-"`
}

func (bp BP) Status(ssp ServiceStatusProvider, r rules.Rules) ResultSet {
	rs := ResultSet{
		Kind:     "BP",
		Name:     bp.Name,
		Id:       bp.Id,
		Children: []ResultSet{},
		Vals:     make(map[string]bool),
	}

	ch := make(chan *ResultSet)
	var calcValues []bool
	for _, k := range bp.Kpis {
		go func(k KPI, ssp ServiceStatusProvider, r rules.Rules) {
			childRs := k.Status(ssp, r)
			ch <- &childRs
		}(k, ssp, r)
	}

	for {
		select {
		case childRs := <-ch:
			calcValues = append(calcValues, childRs.Status.Bool())
			rs.Children = append(rs.Children, *childRs)
			if len(calcValues) == len(bp.Kpis) {
				ch = nil
			}
		}
		if ch == nil {
			break
		}
	}

	ok, _ := calculate("AND", calcValues)
	rs.Status = status.FromBool(ok)
	rs.At = time.Now()
	rs.Vals["in_availability"] = bp.Availability.Contains(rs.At)
	return rs
}

type KPI struct {
	Name      string
	Id        string
	Operation string
	Services  []Service
}

func (k KPI) Status(ssp ServiceStatusProvider, r rules.Rules) ResultSet {
	rs := ResultSet{
		Kind:     "KPI",
		Name:     k.Name,
		Id:       k.Id,
		Children: []ResultSet{},
		Vals:     make(map[string]bool),
	}

	ch := make(chan *ResultSet)
	var calcValues []bool
	for _, s := range k.Services {
		go func(s Service, ssp ServiceStatusProvider, r rules.Rules) {
			childRs := s.Status(ssp, r)
			ch <- &childRs
		}(s, ssp, r)
	}

	for {
		select {
		case childRs := <-ch:
			calcValues = append(calcValues, childRs.Status.Bool())
			rs.Children = append(rs.Children, *childRs)
			if len(calcValues) == len(k.Services) {
				ch = nil
			}
		}
		if ch == nil {
			break
		}
	}

	ok, err := calculate(k.Operation, calcValues)
	rs.Status = status.FromBool(ok)
	rs.At = time.Now()
	if err != nil {
		rs.Err = err
		rs.Status = status.Unknown
	}
	return rs
}

type SvcResult struct {
	At   time.Time
	Msg  string
	Vals map[string]bool
}

type Service struct {
	Host    string
	Service string
}

func (s Service) Status(ssp ServiceStatusProvider, r rules.Rules) ResultSet {
	name := fmt.Sprintf("%s!%s", s.Host, s.Service)
	rs := ResultSet{
		Name: name,
		Id:   name,
		Kind: "SVC",
	}
	at, msg, vals, err := ssp.Status(s.Host, s.Service)
	rs.Err = err
	rs.At = at
	rs.Output = msg
	rs.Vals = vals
	status, err := r.Analyze(vals)
	rs.Status = status
	if rs.Err != nil {
		return rs
	}
	return rs
}
