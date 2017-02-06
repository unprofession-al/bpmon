package cmd

import "time"

type BP struct {
	Name             string       `yaml:"name"`
	Id               string       `yaml:"id"`
	Kpis             []KPI        `yaml:"kpis"`
	AvailabilityName string       `yaml:"availability"`
	Availability     Availability `yaml:"-"`
}

func (bp BP) Status(ssp ServiceStatusProvider) ResultSet {
	rs := ResultSet{
		kind:     "BP",
		name:     bp.Name,
		id:       bp.Id,
		children: []ResultSet{},
		vals:     make(map[string]bool),
	}

	ch := make(chan *ResultSet)
	var calcValues []bool
	for _, k := range bp.Kpis {
		go func(k KPI, ssp ServiceStatusProvider) {
			childRs := k.Status(ssp)
			ch <- &childRs
		}(k, ssp)
	}

	for {
		select {
		case childRs := <-ch:
			calcValues = append(calcValues, childRs.status.toBool())
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
	rs.vals["in_availability"] = bp.Availability.Contains(rs.at)
	if err != nil {
		rs.err = err
		rs.status = StatusUnknown
	}
	return rs
}
