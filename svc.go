package bpmon

import (
	"fmt"
	"time"

	"github.com/unprofession-al/bpmon/rules"
)

type SvcResult struct {
	At   time.Time
	Msg  string
	Vals map[string]bool
}

type ServiceStatusProvider interface {
	Status(string, string) (time.Time, string, map[string]bool, error)
	Values() []string
	DefaultRules() rules.Rules
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
