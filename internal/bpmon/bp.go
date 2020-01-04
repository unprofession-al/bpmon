package bpmon

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/unprofession-al/bpmon/internal/availabilities"
	"github.com/unprofession-al/bpmon/internal/checker"
	"github.com/unprofession-al/bpmon/internal/math"
	"github.com/unprofession-al/bpmon/internal/rules"
	"github.com/unprofession-al/bpmon/internal/status"
	"github.com/unprofession-al/bpmon/internal/store"
	yaml "gopkg.in/yaml.v2"
)

func LoadBP(bpPath string, bpPattern string, a availabilities.Availabilities, gr []string) (BusinessProcesses, error) {
	var bps BusinessProcesses
	if bpPath == "" {
		return bps, nil
	}
	files, err := ioutil.ReadDir(bpPath)
	if err != nil {
		return bps, fmt.Errorf("error while reading business configuration files from '%s': %s", bpPath, err.Error())
	}
	for _, f := range files {
		match, err := filepath.Match(bpPattern, f.Name())
		if err != nil {
			return bps, fmt.Errorf("error while matching file pattern '%s' in '%s': %s", bpPattern, bpPath, err.Error())
		}
		if !match {
			continue
		}
		file, err := ioutil.ReadFile(bpPath + "/" + f.Name())
		if err != nil {
			return bps, fmt.Errorf("error while reading business process %s/%s: %s", bpPath, f.Name(), err.Error())
		}
		bp, err := parseBP(file, a, gr)
		if err != nil {
			return bps, fmt.Errorf("error while parsing business process%s/%s: %s", bpPath, f.Name(), err.Error())
		}
		bps = append(bps, bp)
	}

	return bps, nil
}

func parseBP(bpconf []byte, a availabilities.Availabilities, gr []string) (BP, error) {
	bp := BP{}
	err := yaml.Unmarshal(bpconf, &bp)
	if err != nil {
		return bp, fmt.Errorf("error while parsing: %s", err.Error())
	}

	if bp.AvailabilityName == "" {
		return bp, fmt.Errorf("there is no availability defined in business process config")
	}

	availability, ok := a[bp.AvailabilityName]
	if !ok {
		return bp, fmt.Errorf("the availability referenced '%s' does not exist", bp.AvailabilityName)
	}
	bp.Availability = availability

	if len(gr) > 0 {
		bp.Recipients = append(bp.Recipients, gr...)
	}

	return bp, nil
}

// BusinessProcesses keepes a list of BusinessProcess.
type BusinessProcesses []BP

func (bps BusinessProcesses) GenerateRecipientHashes(pepper string) map[string]string {
	recipientList := make(map[string]struct{})
	for _, bp := range bps {
		for _, recipient := range bp.Recipients {
			recipientList[recipient] = struct{}{}
		}
	}

	recipients := make(map[string]string)
	for recipient := range recipientList {
		toHash := []byte(recipient + ":::" + pepper)
		sum := fmt.Sprintf("%x", sha256.Sum256(toHash))
		recipients[sum] = recipient
	}

	return recipients
}

func (bps BusinessProcesses) GetByRecipients(recipients []string) BusinessProcesses {
	var out BusinessProcesses
	for _, bp := range bps {
		for _, r := range bp.Recipients {
			for _, recipient := range recipients {
				if recipient == r {
					out = append(out, bp)
				}
			}
		}
	}
	return out
}

type BP struct {
	Name             string                      `yaml:"name"`
	ID               string                      `yaml:"id"`
	Kpis             []KPI                       `yaml:"kpis"`
	AvailabilityName string                      `yaml:"availability"`
	Availability     availabilities.Availability `yaml:"-"`
	Responsible      string                      `yaml:"responsible"`
	Recipients       []string                    `yaml:"recipients"`
}

func (bp BP) Status(chk checker.Checker, pp store.Accessor, r rules.Rules) store.ResultSet {
	rs := store.ResultSet{
		Responsible: bp.Responsible,
		Name:        bp.Name,
		ID:          bp.ID,
		Children:    []*store.ResultSet{},
		Vals:        make(map[string]bool),
		Tags:        map[store.Kind]string{store.KindBusinessProcess: bp.ID},
	}

	ch := make(chan *store.ResultSet)
	var calcValues []bool
	for _, k := range bp.Kpis {
		if k.Responsible == "" {
			k.Responsible = bp.Responsible
		}
		go func(k KPI, parentTags map[store.Kind]string, chk checker.Checker, pp store.Accessor, r rules.Rules) {
			childRs := k.Status(rs.Tags, chk, pp, r)
			ch <- &childRs
		}(k, rs.Tags, chk, pp, r)
	}

	for {
		childRs := <-ch
		calcValues = append(calcValues, childRs.Status.Bool())
		rs.Children = append(rs.Children, childRs)
		if len(calcValues) == len(bp.Kpis) {
			ch = nil
		}
		if ch == nil {
			break
		}
	}

	ok, _ := math.Calculate("AND", calcValues)
	rs.Status = status.FromBool(ok)
	rs.Was = status.StatusUnknown
	rs.StatusChanged = false
	rs.Start = time.Now()
	rs.Vals["in_availability"] = bp.Availability.Contains(rs.Start)
	return rs
}

type KPI struct {
	Name        string    `yaml:"name"`
	ID          string    `yaml:"id"`
	Operation   string    `yaml:"operation"`
	Services    []Service `yaml:"services"`
	Responsible string    `yaml:"responsible"`
}

func (k KPI) Status(parentTags map[store.Kind]string, chk checker.Checker, pp store.Accessor, r rules.Rules) store.ResultSet {
	tags := make(map[store.Kind]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[store.KindKeyPerformanceIndicator] = k.ID

	rs := store.ResultSet{
		Responsible: k.Responsible,
		Name:        k.Name,
		ID:          k.ID,
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
		go func(s Service, parentTags map[store.Kind]string, chk checker.Checker, pp store.Accessor, r rules.Rules) {
			childRs := s.Status(rs.Tags, chk, pp, r)
			ch <- &childRs
		}(s, rs.Tags, chk, pp, r)
	}

	for {
		childRs := <-ch
		calcValues = append(calcValues, childRs.Status.Bool())
		rs.Children = append(rs.Children, childRs)
		if len(calcValues) == len(k.Services) {
			ch = nil
		}
		if ch == nil {
			break
		}
	}

	ok, err := math.Calculate(k.Operation, calcValues)
	rs.Status = status.FromBool(ok)
	rs.Was = status.StatusUnknown
	rs.StatusChanged = false
	rs.Start = time.Now()
	if err != nil {
		rs.Err = err
		rs.Status = status.StatusUnknown
	}
	return rs
}

type Service struct {
	Host        string `yaml:"host"`
	Service     string `yaml:"service"`
	Responsible string `yaml:"responsible"`
}

func (s Service) Status(parentTags map[store.Kind]string, chk checker.Checker, pp store.Accessor, r rules.Rules) store.ResultSet {
	name := fmt.Sprintf("%s!%s", s.Host, s.Service)

	tags := make(map[store.Kind]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[store.KindService] = name

	rs := store.ResultSet{
		Name:        name,
		Responsible: s.Responsible,
		ID:          name,
		Tags:        tags,
	}
	result := chk.Status(s.Host, s.Service)
	rs.Err = result.Error
	rs.Start = result.Timestamp
	rs.AppendOutput(result.Message)
	rs.Vals = result.Values
	st, _ := r.Analyze(result.Values)
	rs.Status = st
	rs.Was = status.StatusUnknown
	rs.StatusChanged = false
	return rs
}
