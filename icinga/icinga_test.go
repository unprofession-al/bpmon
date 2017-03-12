package icinga

import (
	"encoding/json"
	"errors"
	"testing"
)

type testset struct {
	host     string
	service  string
	output   string
	result   map[string]bool
	response []byte
}

var TestSets = []testset{
	{
		host:    "testhost",
		service: "testservice",
		output:  "ok",
		response: []byte(`
{
  "results": [
    {
      "attrs": {
        "acknowledgement": 0,
        "last_check_result": {
          "output": "ok",
          "state": 0
        },
        "downtime_depth": 0,
        "last_check": 1488793396.958048
      }
    }
  ]
}
		`),
	},
}

type IcingaMock struct {
	endpoints []testset
}

func (i IcingaMock) Fetch(host, service string) (serviceStatusResponse, error) {
	var response serviceStatusResponse
	for _, ep := range i.endpoints {
		if ep.host == host && ep.service == service {
			err := json.Unmarshal(ep.response, &response)
			return response, err
		}
	}
	return response, errors.New("Service not found")
}

func TestStatusInterpreter(t *testing.T) {
	i := Icinga{fecher: IcingaMock{endpoints: TestSets}}
	for _, test := range TestSets {
		_, output, _, _ := i.Status(test.host, test.service)
		if output != test.output {
			t.Errorf("Failed")
		}
	}
}
