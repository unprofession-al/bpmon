package bpmon

import (
	"errors"
	"testing"
	"time"

	"github.com/unprofession-al/bpmon/persistence"
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
	pp := PPMock{}
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
	vals        map[string]bool
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
	pp := PPMock{}
	chk := CheckMock{}
	for _, s := range SvcTestSets {
		rs := s.svc.Status(chk, pp, chk.DefaultRules())
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

type PPMock struct{}

func (pp PPMock) GetOne(fields []string, from string, where []string, additional string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	status := 0

	if len(fields) < 1 || len(where) < 1 {
		return out, errors.New("Error occured")
	}
	switch where[0] {
	case "ok":
		status = 0
	case "critical":
		status = 1
	case "error":
		return out, errors.New("Error occured")
	default:
		status = 2
	}

	for _, field := range fields {
		out[field] = status
	}
	return out, nil
}

func (pp PPMock) GetAll(fields []string, from string, where []string, additional string) ([]map[string]interface{}, error) {
	var out []map[string]interface{}

	testset := []string{"foo", "bar", "bla"}
	for _, test := range testset {
		set := make(map[string]interface{})
		for _, field := range fields {
			set[field] = test
		}
		out = append(out, set)
	}
	return out, nil
}
func (pp PPMock) Write(p []persistence.Point) error {
	return nil
}
