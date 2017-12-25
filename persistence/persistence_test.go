package persistence

import "errors"

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
