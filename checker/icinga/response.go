package icinga

import (
	"fmt"
	"strconv"
	"time"
)

// Response describes the results returned by the Icinga2
// API when a service status is requested.
type Response struct {
	Results []Result `json:"results"`
}

// Result is a part of the Icinga2 API response reciefed when a service
// status is requested.
type Result struct {
	Attrs Attrs  `json:"attrs"`
	Name  string `json:"name"`
}

// Attrs is a part of the Icinga2 API response reciefed when a service
// status is requested.
type Attrs struct {
	Acknowledgement float64         `json:"acknowledgement"`
	LastCheckResult LastCheckResult `json:"last_check_result"`
	LastCheck       Timestamp       `json:"last_check"`
	DowntimeDepth   float64         `json:"downtime_depth"`
}

// LastCheckResult is a part of the Icinga2 API response reciefed when a
// service status is requested.
type LastCheckResult struct {
	State  float64 `json:"state"`
	Output string  `json:"output"`
}

const (
	statusOK = iota
	statusWarn
	statusCritical
	statusUnknown
)

// Timestamp is used do ensure that malshaling/unmarshaling of timestamps
// works correctly.
type Timestamp time.Time

// MarshalJSON implements the Marshaler interfage.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

// UnmarshalJSON implements the Unmarshaler interfage.
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	tsString, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}

	ts := time.Unix(int64(tsString), 0)
	*t = Timestamp(ts)

	return nil
}
