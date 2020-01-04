package store

import (
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/internal/status"
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
		} else {
			rs.Was = status.StatusUnknown
			rs.WasChecked = true
			if rs.Status != rs.Was {
				rs.StatusChanged = true
			}
			rs.AppendOutput("Error occurred while AddPreviousStatus: " + err.Error())
		}
	}

	for _, childRs := range rs.Children {
		childRs.AddPreviousStatus(pp, saveOK)
	}
}

// AppendOutput appends another string to the Output field and delimits outputs
// with a ` | `.
func (rs *ResultSet) AppendOutput(output string) {
	if rs.Output != "" {
		rs.Output = rs.Output + " | "
	}
	rs.Output = rs.Output + output
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(a, b) {
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
