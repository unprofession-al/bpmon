package main

import (
	"fmt"
	"strconv"
	"time"
)

const (
	IcingaStatusOK = iota
	IcingaStatusWarn
	IcingaStatusCritical
	IcingaStatusUnknown
)

// type IcingaStatusResult describes the results returned by the icinga
// api when a service status is requested.
type IcingaStatusResponse struct {
	Results []IcingaStatusResult `json:"results"`
}

type IcingaStatusResult struct {
	Attrs       IcingaStatusAttrs `json:"attrs"`
	Name        string            `json:"name"`
	HostName    string            `json:"-"`
	ServiceName string            `json:"-"`
}

type IcingaStatusAttrs struct {
	Acknowledgement  float64                     `json:"acknowledgement"`
	LastCheckResults IcingaStatusLastCheckResult `json:"last_check_result"`
	LastCheck        Timestamp                   `json:"last_check"`
	DowntimeDepth    float64                     `json:"downtime_depth"`
}

type IcingaStatusLastCheckResult struct {
	State  float64 `json:"state"`
	Output string  `json:"output"`
}

type Timestamp struct {
	time.Time
}

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := t.Time.Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}

	t.Time = time.Unix(int64(ts), 0)

	return nil
}
