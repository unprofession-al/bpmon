package cmd

import (
	"fmt"
	"time"
)

type SvcResult struct {
	At   time.Time
	Msg  string
	Vals map[string]bool
}

type Resolution struct {
	Vals map[string]bool
	Is   Status
}

type ServiceStatusProvider interface {
	Status(Service) (SvcResult, error)
	Values() []string
	Analyze(SvcResult) (Status, error)
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
	status, err := ssp.Analyze(result)
	rs.Status = status
	if rs.Err != nil {
		return rs
	}
	rs.Err = err
	return rs
}
