package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type BusinessProcess struct {
	Name             string       `yaml:"name"`
	Id               string       `yaml:"id"`
	Kpis             []kpi        `yaml:"kpis"`
	AvailabilityName string       `yaml:"availability"`
	Availability     Availability `yaml:"-"`
}

type ServiceStatusProvider interface {
	ServiceStatus(Service) (bool, time.Time, bool, bool, string, error)
}

func (bp BusinessProcess) Status(ssp ServiceStatusProvider) ResultSet {
	rs := ResultSet{
		kind:     "BP",
		name:     bp.Name,
		id:       bp.Id,
		children: []ResultSet{},
	}

	ch := make(chan *ResultSet)
	var calcValues []bool
	for _, k := range bp.Kpis {
		go func(k kpi, ssp ServiceStatusProvider) {
			childRs := k.Status(ssp)
			ch <- &childRs
		}(k, ssp)
	}

	for {
		select {
		case childRs := <-ch:
			calcValues = append(calcValues, childRs.considerHealthy())
			rs.children = append(rs.children, *childRs)
			if len(calcValues) == len(bp.Kpis) {
				ch = nil
			}
		}
		if ch == nil {
			break
		}
	}

	ok, err := calculate("AND", calcValues)
	rs.status = boolAsStatus(ok)
	rs.at = time.Now()
	rs.inAvailability = bp.Availability.Contains(rs.at)
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
		kind:           "KPI",
		name:           k.Name,
		id:             k.Id,
		children:       []ResultSet{},
		inDowntime:     false,
		inAvailability: true,
	}

	ch := make(chan *ResultSet)
	var calcValues []bool
	for _, s := range k.Services {
		go func(s Service, ssp ServiceStatusProvider) {
			childRs := s.Status(ssp)
			ch <- &childRs
		}(s, ssp)
	}

	for {
		select {
		case childRs := <-ch:
			calcValues = append(calcValues, childRs.considerHealthy())
			rs.children = append(rs.children, *childRs)
			if len(calcValues) == len(k.Services) {
				ch = nil
			}
		}
		if ch == nil {
			break
		}
	}

	ok, err := calculate(k.Operation, calcValues)
	rs.status = boolAsStatus(ok)
	rs.at = time.Now()
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
		name:           name,
		id:             name,
		kind:           "SVC",
		inAvailability: true,
	}
	ok, at, inDowntime, acknowledged, output, err := ssp.ServiceStatus(s)
	rs.err = err
	rs.inDowntime = inDowntime
	rs.at = at
	rs.acknowledged = acknowledged
	rs.output = output
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
	name           string
	id             string
	kind           string
	at             time.Time
	inDowntime     bool
	acknowledged   bool
	inAvailability bool
	status         status
	err            error
	output         string
	children       []ResultSet
}

func (rs ResultSet) PrettyPrint(level int) string {
	ident := strings.Repeat("   ", level)
	out := rs.status.Colorize(fmt.Sprintf("%s%s %s is %v", ident, rs.kind, rs.name, rs.status))

	ident = strings.Repeat("   ", level+1)
	out += fmt.Sprintf("\n%sMeasured at: %v", ident, rs.at)
	if rs.err != nil {
		out += fmt.Sprintf("\n%sError occured: %s", ident, rs.err.Error())
	}
	if !rs.inAvailability {
		out += fmt.Sprintf("\n%sMeasured outside of required availability", ident)
	}
	if rs.inDowntime {
		out += fmt.Sprintf("\n%sMeasured in Scheduled Downtime", ident)
	}
	if rs.acknowledged {
		out += fmt.Sprintf("\n%sIs acknowledged", ident)
	}
	if rs.status == StatusNOK && rs.output != "" {
		out += fmt.Sprintf("\n%sMessage from Monitoring: %s", ident, rs.output)
	}
	out += "\n"
	for _, childRs := range rs.children {
		out += childRs.PrettyPrint(level + 1)
	}
	return out
}

func (rs ResultSet) AsInflux(parentTags map[string]string, saveOK []string) []Point {
	var out []Point

	tags := make(map[string]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[rs.kind] = rs.id

	if rs.status != StatusOK || stringInSlice(rs.kind, saveOK) {
		fields := map[string]interface{}{
			"status":         rs.status.toInt(),
			"acknowledged":   strconv.FormatBool(rs.acknowledged),
			"inAvailability": strconv.FormatBool(rs.inAvailability),
			"inDowntime":     strconv.FormatBool(rs.inDowntime),
		}
		if rs.output != "" {
			fields["output"] = fmt.Sprintf("Output: %s", rs.output)
		}
		if rs.err != nil {
			fields["err"] = fmt.Sprintf("Error: %s", rs.err.Error())
		}
		pt := Point{
			Timestamp: rs.at,
			Series:    rs.kind,
			Tags:      tags,
			Fields:    fields,
		}
		out = append(out, pt)
	}

	for _, childRs := range rs.children {
		out = append(out, childRs.AsInflux(tags, saveOK)...)
	}
	return out
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToUpper(b) == strings.ToUpper(a) {
			return true
		}
	}
	return false
}

// considerHealthy returns a boolean representation of the status. 'true' means
// that either the status is fine, unknown or the check was in a scheduled
// downtime; 'false' means that the check was negative
func (rs ResultSet) considerHealthy() bool {
	if rs.status == StatusNOK && !rs.inDowntime {
		return false
	}
	return true
}
