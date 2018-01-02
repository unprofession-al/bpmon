package influx

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

type point struct {
	Timestamp time.Time              `json:"timestamp"`
	Series    string                 `json:"series"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

func (i Influx) asPoints(rs *store.ResultSet) []point {
	var out []point

	if rs.Status != status.OK || stringInSlice(rs.Kind(), i.saveOK) {
		fields := map[string]interface{}{
			"status":    rs.Status.Int(),
			"annotated": rs.Annotated,
		}
		for key, value := range rs.Vals {
			fields[key] = value
		}
		if rs.Output != "" {
			fields["output"] = fmt.Sprintf("Output: %s", rs.Output)
		}
		if rs.Err != nil {
			fields["err"] = fmt.Sprintf("Error: %s", rs.Err.Error())
		}
		if rs.WasChecked {
			fields["was"] = rs.Was.Int()
			fields["changed"] = rs.StatusChanged
		}
		if rs.Annotation != "" {
			fields["annotation"] = rs.Annotation
		}

		pt := point{
			Timestamp: rs.Start,
			Series:    rs.Kind(),
			Tags:      rs.Tags,
			Fields:    fields,
		}
		out = append(out, pt)
	}

	for _, childRs := range rs.Children {
		out = append(out, i.asPoints(childRs)...)
	}
	return out
}

func (i Influx) asResultSet(data map[string]interface{}) (store.ResultSet, error) {
	var err error
	var ok bool
	out := store.ResultSet{
		Tags: make(map[string]string),
		Vals: make(map[string]bool),
	}
	for k, v := range data {
		if v != nil {
			switch k {
			case timefield:
				out.Start, err = time.Parse(time.RFC3339, v.(string))
				if err != nil {
					return out, err
				}
			case store.IdentifierBusinessProcess:
				out.Tags[store.IdentifierBusinessProcess] = v.(string)
			case store.IdentifierKeyPerformanceIndicator:
				out.Tags[store.IdentifierKeyPerformanceIndicator] = v.(string)
			case store.IdentifierService:
				out.Tags[store.IdentifierService] = v.(string)
			case "err":
				out.Err = errors.New(v.(string))
			case "output":
				out.Output = v.(string)
			case "annotation":
				out.Annotation = v.(string)
			case "annotated":
				out.Annotated, ok = v.(bool)
				if !ok {
					return out, fmt.Errorf("Could not convert %v (type %s) to bool for 'annotated'", v, reflect.TypeOf(v))
				}
			case "changed":
				out.StatusChanged, ok = v.(bool)
				if !ok {
					return out, fmt.Errorf("Could not convert %v (type %s) to bool for 'changed'", v, reflect.TypeOf(v))
				}
			case "status":
				raw, err := v.(json.Number).Int64()
				if err != nil {
					return out, err
				}
				out.Status, err = status.FromInt64(raw)
				if err != nil {
					return out, err
				}
			case "was":
				out.WasChecked = true
				raw, err := v.(json.Number).Int64()
				if err != nil {
					return out, err
				}
				out.Was, err = status.FromInt64(raw)
				if err != nil {
					return out, err
				}
			case "end":
			case "name":
			case "responsible":
			case "children":
			default:
				out.Vals[k], ok = v.(bool)
				if !ok {
					return out, fmt.Errorf("Could not convert %v (type %s) to bool for '%s'", v, reflect.TypeOf(v), k)
				}
			}
		}
	}
	out.ID = out.Tags[out.Kind()]
	return out, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToUpper(b) == strings.ToUpper(a) {
			return true
		}
	}
	return false
}
