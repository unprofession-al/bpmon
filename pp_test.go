package bpmon

import "errors"

type PPMock struct{}

func (pp PPMock) GetOne(query string) (interface{}, error) {
	switch query {
	case "ok":
		return "0", nil
	case "critical":
		return "1", nil
	case "error":
		return "2", errors.New("Error occured")
	default:
		return "2", nil
	}
}
