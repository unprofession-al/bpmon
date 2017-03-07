package bpmon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

type BusinessProcesses []BP

type BP struct {
	Name             string       `yaml:"name"`
	Id               string       `yaml:"id"`
	Kpis             []KPI        `yaml:"kpis"`
	AvailabilityName string       `yaml:"availability"`
	Availability     Availability `yaml:"-"`
}

func readBPs(bpPath, bpPattern string, a Availabilities) (BusinessProcesses, error) {
	bps := BusinessProcesses{}

	files, err := ioutil.ReadDir(bpPath)
	if err != nil {
		return bps, err
	}

	for _, f := range files {
		match, err := filepath.Match(bpPattern, f.Name())
		if err != nil {
			return bps, err
		}
		if !match {
			continue
		}
		bp := BP{}
		file, err := ioutil.ReadFile(bpPath + "/" + f.Name())
		if err != nil {
			err = errors.New(fmt.Sprintf("Error while reading %s/%s: %s", bpPath, f.Name(), err.Error()))
			return bps, err
		}

		err = yaml.Unmarshal(file, &bp)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error while parsing %s/%s: %s", bpPath, f.Name(), err.Error()))
			return bps, err
		}

		if bp.AvailabilityName == "" {
			err = errors.New(fmt.Sprintf("There is no availability defined in %s/%s", bpPath, f.Name()))
			return bps, err
		}

		availability, ok := a[bp.AvailabilityName]
		if !ok {
			err = errors.New(fmt.Sprintf("The availability '%s' referenced in '%s/%s' does not exist", bp.AvailabilityName, bpPath, f.Name()))
			return bps, err
		}
		bp.Availability = availability

		bps = append(bps, bp)
	}
	return bps, nil
}

func (bp BP) Status(ssp ServiceStatusProvider) ResultSet {
	rs := ResultSet{
		Kind:     "BP",
		Name:     bp.Name,
		Id:       bp.Id,
		Children: []ResultSet{},
		Vals:     make(map[string]bool),
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
			calcValues = append(calcValues, childRs.Status.toBool())
			rs.Children = append(rs.Children, *childRs)
			if len(calcValues) == len(bp.Kpis) {
				ch = nil
			}
		}
		if ch == nil {
			break
		}
	}

	ok, err := calculate("AND", calcValues)
	rs.Status = boolAsStatus(ok)
	rs.At = time.Now()
	rs.Vals["in_availability"] = bp.Availability.Contains(rs.At)
	if err != nil {
		rs.Err = err
		rs.Status = StatusUnknown
	}
	return rs
}
