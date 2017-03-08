package bpmon

import (
	"time"

	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
)

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
			calcValues = append(calcValues, childRs.Status.ToBool())
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
	rs.Status = status.BoolAsStatus(ok)
	rs.At = time.Now()
	if err != nil {
		rs.Err = err
		rs.Status = status.Unknown
	}
	return rs
}
