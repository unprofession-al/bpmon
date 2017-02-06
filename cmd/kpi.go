package cmd

import "time"

type KPI struct {
	Name      string
	Id        string
	Operation string
	Services  []Service
}

func (k KPI) Status(ssp ServiceStatusProvider) ResultSet {
	rs := ResultSet{
		kind:     "KPI",
		name:     k.Name,
		id:       k.Id,
		children: []ResultSet{},
		vals:     make(map[string]bool),
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
			calcValues = append(calcValues, childRs.status.toBool())
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
