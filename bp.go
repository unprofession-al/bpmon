package bpmon

import (
	"fmt"
	"time"

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/store"
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
	Responsible      string       `yaml:"responsible"`
}

func (bp BP) Status(chk checker.Checker, pp store.Store, r rules.Rules) store.ResultSet {
	rs := store.ResultSet{
		Responsible: bp.Responsible,
		Name:        bp.Name,
		Id:          bp.Id,
		Children:    []*store.ResultSet{},
		Vals:        make(map[string]bool),
		Tags:        map[string]string{store.IdentifierBusinessProcess: bp.Id},
	}

	ch := make(chan *store.ResultSet)
	var calcValues []bool
	for _, k := range bp.Kpis {
		if k.Responsible == "" {
			k.Responsible = bp.Responsible
		}
		go func(k KPI, parentTags map[string]string, chk checker.Checker, pp store.Store, r rules.Rules) {
			childRs := k.Status(rs.Tags, chk, pp, r)
			ch <- &childRs
		}(k, rs.Tags, chk, pp, r)
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
	Name        string    `yaml:"name"`
	Id          string    `yaml:"id"`
	Operation   string    `yaml:"operation"`
	Services    []Service `yaml:"services"`
	Responsible string    `yaml:"responsible"`
}

func (k KPI) Status(parentTags map[string]string, chk checker.Checker, pp store.Store, r rules.Rules) store.ResultSet {
	tags := make(map[string]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[store.IdentifierKeyPerformanceIndicator] = k.Id

	rs := store.ResultSet{
		Responsible: k.Responsible,
		Name:        k.Name,
		Id:          k.Id,
		Children:    []*store.ResultSet{},
		Vals:        make(map[string]bool),
		Tags:        tags,
	}

	ch := make(chan *store.ResultSet)
	var calcValues []bool
	for _, s := range k.Services {
		if s.Responsible == "" {
			s.Responsible = k.Responsible
		}
		go func(s Service, parentTags map[string]string, chk checker.Checker, pp store.Store, r rules.Rules) {
			childRs := s.Status(rs.Tags, chk, pp, r)
			ch <- &childRs
		}(s, rs.Tags, chk, pp, r)
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

type Service struct {
	Host        string `yaml:"host"`
	Service     string `yaml:"service"`
	Responsible string `yaml:"responsible"`
}

func (s Service) Status(parentTags map[string]string, chk checker.Checker, pp store.Store, r rules.Rules) store.ResultSet {
	name := fmt.Sprintf("%s!%s", s.Host, s.Service)

	tags := make(map[string]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[store.IdentifierService] = name

	rs := store.ResultSet{
		Name:        name,
		Responsible: s.Responsible,
		Id:          name,
		Tags:        tags,
	}
	result := chk.Status(s.Host, s.Service)
	rs.Err = result.Error
	rs.At = result.Timestamp
	rs.Output = result.Message
	rs.Vals = result.Values
	st, _ := r.Analyze(result.Values)
	rs.Status = st
	rs.Was = status.Unknown
	rs.StatusChanged = false
	return rs
}
