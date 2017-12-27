package store

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mgutz/ansi"
	"github.com/unprofession-al/bpmon/status"
)

type ResultSet struct {
	Name          string
	Id            string
	Tags          map[string]string
	At            time.Time
	Vals          map[string]bool
	Status        status.Status
	Was           status.Status
	WasChecked    bool
	Annotated     bool
	Annotation    string
	StatusChanged bool
	Err           error
	Output        string
	Responsible   string
	Children      []*ResultSet
}

const (
	IdentifierBusinessProcess         = "BP"
	IdentifierKeyPerformanceIndicator = "KPI"
	IdentifierService                 = "SVC"
)

func (rs ResultSet) Kind() string {
	kind := "UNKNOWN"
	if _, ok := rs.Tags[IdentifierBusinessProcess]; ok {
		kind = IdentifierBusinessProcess
	}
	if _, ok := rs.Tags[IdentifierKeyPerformanceIndicator]; ok {
		kind = IdentifierKeyPerformanceIndicator
	}
	if _, ok := rs.Tags[IdentifierService]; ok {
		kind = IdentifierService
	}
	return kind
}

func (rs ResultSet) PrettyPrint(level int, ts bool, vals bool, resp bool) string {
	ident := strings.Repeat("   ", level)
	out := rs.Status.Colorize(fmt.Sprintf("%s%s %s is %v", ident, rs.Kind(), rs.Name, rs.Status))
	if rs.WasChecked {
		out += rs.Status.Colorize(fmt.Sprintf(" (was %v)", rs.Was))
	}

	ident = strings.Repeat("   ", level+1)
	if ts {
		timestamp := rs.At.Format("2006-01-02 15:04:05")
		out += fmt.Sprintf(" (%s)", timestamp)
	}
	if resp && rs.Responsible != "" {
		out += fmt.Sprintf(" (Responsible: %s)", rs.Responsible)
	}
	if rs.Err != nil {
		out += fmt.Sprintf("\n%sError occured: %s", ident, rs.Err.Error())
	}
	if rs.Status == status.Nok && rs.Output != "" {
		out += fmt.Sprintf("\n%sMessage from Monitoring: %s", ident, rs.Output)
	}
	if vals && len(rs.Vals) > 0 {
		out += fmt.Sprintf("\n%sValues:", ident)
		var keys []string
		for k := range rs.Vals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := rs.Vals[key]
			out += " "
			if !value {
				out += ansi.Color(key, "magenta+s")
			} else {
				out += ansi.Color(key, "green")
			}
		}
	}
	out += "\n"
	for _, childRs := range rs.Children {
		out += childRs.PrettyPrint(level+1, ts, vals, resp)
	}
	return out
}

func (rs ResultSet) StripByStatus(s []status.Status) (ResultSet, bool) {
	setOut := rs
	keep := true
	for _, status := range s {
		if rs.Status == status {
			keep = false
			break
		}
	}
	if keep {
		var children []*ResultSet
		for _, child := range rs.Children {
			set, stripped := child.StripByStatus(s)
			if !stripped {
				children = append(children, &set)
			}
		}
		setOut.Children = children
	}
	return setOut, !keep
}

func (rs *ResultSet) AddPreviousStatus(pp Store, saveOK []string) {
	if stringInSlice(rs.Kind(), saveOK) {
		latest, err := pp.GetLatest(*rs)
		if err == nil {
			rs.Was = latest.Was
			rs.WasChecked = true
			if rs.Status != rs.Was {
				rs.StatusChanged = true
			}
		}
	}

	for _, childRs := range rs.Children {
		childRs.AddPreviousStatus(pp, saveOK)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToUpper(b) == strings.ToUpper(a) {
			return true
		}
	}
	return false
}
