package bpmon

import (
	"fmt"
	"time"
)

type SvcResult struct {
	At   time.Time
	Msg  string
	Vals map[string]bool
}

type ServiceStatusProvider interface {
	Status(Service) (SvcResult, error)
	Values() []string
	Rules() Rules
}

type Service struct {
	Host    string
	Service string
}

func (s Service) Status(ssp ServiceStatusProvider) ResultSet {
	name := fmt.Sprintf("%s!%s", s.Host, s.Service)
	rs := ResultSet{
		Name: name,
		Id:   name,
		Kind: "SVC",
	}
	result, err := ssp.Status(s)
	rs.Err = err
	rs.At = result.At
	rs.Output = result.Msg
	rs.Vals = result.Vals
	status, err := ssp.Rules().Analyze(result)
	rs.Status = status
	if rs.Err != nil {
		return rs
	}
	return rs
}
