package bpmon

import (
	"fmt"
	"time"

	"github.com/unprofession-al/bpmon/persistence"
	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
)

const (
	IdentifierBusinessProcess         = "BP"
	IdentifierKeyPerformanceIndicator = "KPI"
	IdentifierService                 = "SVC"
)

func getKind(spec map[string]string) string {
	kind := "UNKNOWN"
	if _, ok := spec[IdentifierBusinessProcess]; ok {
		kind = IdentifierBusinessProcess
	}
	if _, ok := spec[IdentifierKeyPerformanceIndicator]; ok {
		kind = IdentifierKeyPerformanceIndicator
	}
	if _, ok := spec[IdentifierService]; ok {
		kind = IdentifierService
	}
	return kind
}

type BusinessProcesses []BP

type BP struct {
	Name             string       `yaml:"name"`
	Id               string       `yaml:"id"`
	Kpis             []KPI        `yaml:"kpis"`
	AvailabilityName string       `yaml:"availability"`
	Availability     Availability `yaml:"-"`
	Responsible      string       `yaml:"responsible"`
}

func (bp BP) Status(ssp ServiceStatusProvider, pp persistence.Persistence, r rules.Rules) persistence.ResultSet {
	rs := persistence.ResultSet{
		Kind:        IdentifierBusinessProcess,
		Responsible: bp.Responsible,
		Name:        bp.Name,
		Id:          bp.Id,
		Children:    []*persistence.ResultSet{},
		Vals:        make(map[string]bool),
	}

	ch := make(chan *persistence.ResultSet)
	var calcValues []bool
	for _, k := range bp.Kpis {
		if k.Responsible == "" {
			k.Responsible = bp.Responsible
		}
		go func(k KPI, ssp ServiceStatusProvider, pp persistence.Persistence, r rules.Rules) {
			childRs := k.Status(ssp, pp, r)
			ch <- &childRs
		}(k, ssp, pp, r)
	}

	for {
		select {
		case childRs := <-ch:
			calcValues = append(calcValues, childRs.Status.Bool())
			rs.Children = append(rs.Children, childRs)
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
	rs.Was = status.Unknown
	rs.StatusChanged = false
	rs.At = time.Now()
	rs.Vals["in_availability"] = bp.Availability.Contains(rs.At)
	return rs
}

type KPI struct {
	Name        string
	Id          string
	Operation   string
	Services    []Service
	Responsible string
}

func (k KPI) Status(ssp ServiceStatusProvider, pp persistence.Persistence, r rules.Rules) persistence.ResultSet {
	rs := persistence.ResultSet{
		Kind:        IdentifierKeyPerformanceIndicator,
		Responsible: k.Responsible,
		Name:        k.Name,
		Id:          k.Id,
		Children:    []*persistence.ResultSet{},
		Vals:        make(map[string]bool),
	}

	ch := make(chan *persistence.ResultSet)
	var calcValues []bool
	for _, s := range k.Services {
		if s.Responsible == "" {
			s.Responsible = k.Responsible
		}
		go func(s Service, ssp ServiceStatusProvider, pp persistence.Persistence, r rules.Rules) {
			childRs := s.Status(ssp, pp, r)
			ch <- &childRs
		}(s, ssp, pp, r)
	}

	for {
		select {
		case childRs := <-ch:
			calcValues = append(calcValues, childRs.Status.Bool())
			rs.Children = append(rs.Children, childRs)
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
	rs.Was = status.Unknown
	rs.StatusChanged = false
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
	Host        string
	Service     string
	Responsible string
}

func (s Service) Status(ssp ServiceStatusProvider, pp persistence.Persistence, r rules.Rules) persistence.ResultSet {
	name := fmt.Sprintf("%s!%s", s.Host, s.Service)
	rs := persistence.ResultSet{
		Name:        name,
		Responsible: s.Responsible,
		Id:          name,
		Kind:        IdentifierService,
	}
	at, msg, vals, err := ssp.Status(s.Host, s.Service)
	rs.Err = err
	rs.At = at
	rs.Output = msg
	rs.Vals = vals
	st, err := r.Analyze(vals)
	rs.Status = st
	rs.Was = status.Unknown
	rs.StatusChanged = false
	if rs.Err != nil {
		return rs
	}
	return rs
}
