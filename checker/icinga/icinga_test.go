package icinga

import (
	"encoding/json"
	"errors"
	"reflect"
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
		host:    "Test Host",
		service: "All Fine",
		output:  "ok",
		result: map[string]bool{
			"ok":                 true,
			"unknown":            false,
			"warn":               false,
			"critical":           false,
			"scheduled_downtime": false,
			"acknowledged":       false,
			"failed":             false,
		},
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
	{
		host:    "Test Host",
		service: "Ack, Downtime, Critical",
		output:  "failed",
		result: map[string]bool{
			"ok":                 false,
			"unknown":            false,
			"warn":               false,
			"critical":           true,
			"scheduled_downtime": true,
			"acknowledged":       true,
			"failed":             false,
		},
		response: []byte(`
{
  "results": [
    {
      "attrs": {
        "acknowledgement": 1,
        "last_check_result": {
          "output": "failed",
          "state": 2
        },
        "downtime_depth": 1,
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

func (i IcingaMock) Fetch(host, service string) (Response, error) {
	var response Response
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
		result := i.Status(test.host, test.service)
		if result.Error != nil {
			t.Errorf("Error returned: %s", result.Error.Error())
		}
		if result.Message != test.output {
			t.Errorf("Failed")
		}
		eq := reflect.DeepEqual(result.Values, test.result)
		if !eq {
			t.Errorf("Results do not match: '%v' vs. '%v'", result.Values, test.result)
		}
	}
}
