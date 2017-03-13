package bpmon

import (
	"testing"

	"github.com/unprofession-al/bpmon/status"
)

type svcTestSet struct {
	svc         Service
	status      status.Status
	vals        map[string]bool
	errExpected bool
}

var SvcTestSets = []svcTestSet{
	{
		svc:         Service{"Host", "good"},
		status:      status.Ok,
		errExpected: false,
	},
	{
		svc:         Service{"Host", "bad"},
		status:      status.Nok,
		errExpected: false,
	},
	{
		svc:         Service{"Host", "unknown"},
		status:      status.Unknown,
		errExpected: false,
	},
	{
		svc:         Service{"Host", "error"},
		status:      status.Ok,
		errExpected: true,
	},
}

func TestServices(t *testing.T) {
	ssp := SSPMock{}
	for _, s := range SvcTestSets {
		rs := s.svc.Status(ssp, ssp.DefaultRules())
		if s.errExpected && rs.Err == nil {
			t.Errorf("Error expected but got nil")
		} else if !s.errExpected && rs.Err != nil {
			t.Errorf("No error expected  but got error: %s", rs.Err.Error())
		}
		if rs.Status != s.status {
			t.Errorf("Expected status to be '%s', got '%s'", s.status, rs.Status)
		}

	}
}
