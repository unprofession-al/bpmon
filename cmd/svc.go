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
		name: name,
		id:   name,
		kind: "SVC",
	}
	result, err := ssp.Status(s)
	rs.err = err
	rs.at = result.At
	rs.output = result.Msg
	rs.vals = result.Vals
	status, _ := ssp.Analyze(result)
	rs.status = status
	return rs
}
