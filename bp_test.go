package bpmon

import (
	"testing"
	"time"

	"github.com/unprofession-al/bpmon/status"
)

type bpTestSet struct {
	bp     BP
	status status.Status
}

var allDayLong = Availability{
	time.Monday:    AvailabilityTime{AllDay: true},
	time.Tuesday:   AvailabilityTime{AllDay: true},
	time.Wednesday: AvailabilityTime{AllDay: true},
	time.Thursday:  AvailabilityTime{AllDay: true},
	time.Friday:    AvailabilityTime{AllDay: true},
	time.Saturday:  AvailabilityTime{AllDay: true},
	time.Sunday:    AvailabilityTime{AllDay: true},
}

var BpTestSets = []bpTestSet{
	{
		bp: BP{
			Name:         "TestBP",
			Id:           "test_bp",
			Availability: allDayLong,
			Kpis: []KPI{
				KPI{
					Name:      "TestKPI",
					Id:        "test_kpi",
					Operation: "OR",
					Services: []Service{
						Service{Host: "Host", Service: "good"},
					},
				},
			},
		},
		status: status.Ok,
	},
}

func TestBusinessProcess(t *testing.T) {
	chk := CheckerMock{}
	pp := PersistenceMock{}
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
		status:      status.Ok,
		errExpected: false,
	},
	{
		svc:         Service{Host: "Host", Service: "bad"},
		status:      status.Nok,
		errExpected: false,
	},
	{
		svc:         Service{Host: "Host", Service: "unknown"},
		status:      status.Unknown,
		errExpected: false,
	},
	{
		svc:         Service{Host: "Host", Service: "error"},
		status:      status.Ok,
		errExpected: true,
	},
}

func TestServices(t *testing.T) {
	pp := PersistenceMock{}
	chk := CheckerMock{}
	parentTags := map[string]string{"BP": "BP", "KPI": "KPI"}
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
