package store

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mgutz/ansi"
	"github.com/unprofession-al/bpmon/status"
)

// ResultSet holds all results of a check. It is also returned by store
// implementations when queries are executed.
type ResultSet struct {
	Name          string
	ID            string
	Start         time.Time
	Tags          map[Kind]string
	Vals          map[string]bool
	Status        status.Status
	Was           status.Status
	WasChecked    bool
	StatusChanged bool
	Annotated     bool
	Annotation    string
	Err           error
	Output        string
	Responsible   string
	Children      []*ResultSet
}

// Kind returns the Kind of the Result set based on its tags.
func (rs ResultSet) Kind() Kind {
	kind := KindUnknown
	if _, ok := rs.Tags[KindBusinessProcess]; ok {
		kind = KindBusinessProcess
	}
	if _, ok := rs.Tags[KindKeyPerformanceIndicator]; ok {
		kind = KindKeyPerformanceIndicator
	}
	if _, ok := rs.Tags[KindService]; ok {
		kind = KindService
	}
	return kind
}

// PrettyPrint returns a string representation of its 'ResultSet', proper formated
// to be printed to STDOUT in a human readable form.
func (rs ResultSet) PrettyPrint(level int, ts bool, vals bool, resp bool) string {
	ident := strings.Repeat("   ", level)
	out := rs.Status.Colorize(fmt.Sprintf("%s%s %s is %v", ident, rs.Kind(), rs.Name, rs.Status))
	if rs.WasChecked {
		out += rs.Status.Colorize(fmt.Sprintf(" (was %v)", rs.Was))
	}

	ident = strings.Repeat("   ", level+1)
	if ts {
		timestamp := rs.Start.Format("2006-01-02 15:04:05")
		out += fmt.Sprintf(" (%s)", timestamp)
	}
	if resp && rs.Responsible != "" {
		out += fmt.Sprintf(" (Responsible: %s)", rs.Responsible)
	}
	if rs.Err != nil {
		out += fmt.Sprintf("\n%sError occured: %s", ident, rs.Err.Error())
	}
	if rs.Status == status.StatusNOK && rs.Output != "" {
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

// FilterByStatus returns a ResultSet that only contains element with one of the
// status passed as argument.
func (rs ResultSet) FilterByStatus(s []status.Status) (ResultSet, bool) {
	setOut := rs
	keep := false
	for _, status := range s {
		if rs.Status == status {
			keep = true
			break
		}
	}
	if keep {
		var children []*ResultSet
		for _, child := range rs.Children {
			set, stripped := child.FilterByStatus(s)
			if !stripped {
				children = append(children, &set)
			}
		}
		setOut.Children = children
	}
	return setOut, !keep
}

// AddPreviousStatus looks up the last status of a 'ResultSet' which was
// persisted to the store and adds 3 additional to the current ResultSet:
//
//		* 'Was' represent the latest status
//		* 'WasChecked' is set to 'true'
//		* 'StatusChanged' is set to 'true' if the last Status and the current status differ
//
// This is only executed if status.OK is peristed for the 'Kind' of the
// 'ResultSet', otherwise the information would not make any sense.
func (rs *ResultSet) AddPreviousStatus(pp Accessor, saveOK []string) {
	if stringInSlice(string(rs.Kind()), saveOK) {
		latest, err := pp.GetLatest(*rs)
		if err == nil {
			rs.Was = latest.Status
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

// Kind describes the type of entity a ResultSet/Event/Store can represent.
type Kind string

// Kind constants describe all possible types that a ResultSet/Event/Store
//can represent.
const (
	// KindBusinessProcess is used for Business Processes
	KindBusinessProcess Kind = "BP"

	// KindKeyPerformanceIndicator is used for Key Performance Indicator
	KindKeyPerformanceIndicator Kind = "KPI"

	// KindService is used for Service
	KindService Kind = "SVC"

	// KindUnknown is used for Unknown types
	KindUnknown Kind = "UNKNOWN"
)

// String implements the Stringer interface.
func (k Kind) String() string {
	return string(k)
}
