package bpmon

import (
	"testing"
	"time"

	"github.com/unprofession-al/bpmon/internal/availabilities"
	"github.com/unprofession-al/bpmon/internal/status"
	"github.com/unprofession-al/bpmon/internal/store"
)

type bpTestSet struct {
	bp     BP
	status status.Status
}

var allDayLong = availabilities.Availability{
	time.Monday:    availabilities.AvailabilityTime{AllDay: true},
	time.Tuesday:   availabilities.AvailabilityTime{AllDay: true},
	time.Wednesday: availabilities.AvailabilityTime{AllDay: true},
	time.Thursday:  availabilities.AvailabilityTime{AllDay: true},
	time.Friday:    availabilities.AvailabilityTime{AllDay: true},
	time.Saturday:  availabilities.AvailabilityTime{AllDay: true},
	time.Sunday:    availabilities.AvailabilityTime{AllDay: true},
}

var BpTestSets = []bpTestSet{
	{
		bp: BP{
			Name:         "TestBP",
			ID:           "test_bp",
			Availability: allDayLong,
			Kpis: []KPI{
				KPI{
					Name:      "TestKPI",
					ID:        "test_kpi",
					Operation: "OR",
					Services: []Service{
						Service{Host: "Host", Service: "good"},
					},
				},
			},
		},
		status: status.StatusOK,
	},
}

func TestBusinessProcess(t *testing.T) {
	chk := CheckerMock{}
	pp := StoreMock{}
	for _, bp := range BpTestSets {
		rs := bp.bp.Status(chk, pp, chk.DefaultRules())
		if rs.Status != bp.status {
			t.Errorf("Expected status to be '%s', got '%s'", bp.status, rs.Status)
		}
	}
}

type svcTestSet struct {
	svc         Service
	status      status.Status
	errExpected bool
}

var SvcTestSets = []svcTestSet{
	{
		svc:         Service{Host: "Host", Service: "good"},
		status:      status.StatusOK,
		errExpected: false,
	},
	{
		svc:         Service{Host: "Host", Service: "bad"},
		status:      status.StatusNOK,
		errExpected: false,
	},
	{
		svc:         Service{Host: "Host", Service: "unknown"},
		status:      status.StatusUnknown,
		errExpected: false,
	},
	{
		svc:         Service{Host: "Host", Service: "error"},
		status:      status.StatusOK,
		errExpected: true,
	},
}

func TestServices(t *testing.T) {
	pp := StoreMock{}
	chk := CheckerMock{}
	parentTags := map[store.Kind]string{store.KindBusinessProcess: "BP", store.KindKeyPerformanceIndicator: "KPI"}
	for _, s := range SvcTestSets {
		rs := s.svc.Status(parentTags, chk, pp, chk.DefaultRules())
		if s.errExpected && rs.Err == nil {
			t.Errorf("Error expected but got nil")
		} else if !s.errExpected && rs.Err != nil {
			t.Errorf("No error expected but got error: %s", rs.Err.Error())
		}
		if rs.Status != s.status {
			t.Errorf("Expected status to be '%s', got '%s'", s.status, rs.Status)
		}
	}
}
