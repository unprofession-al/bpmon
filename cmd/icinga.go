package cmd

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type IcingaConf struct {
	Connection struct {
		Server string
		Port   int
		Pass   string
		User   string
		Proto  string
	}
}

type Icinga struct {
	baseUrl string
	user    string
	pass    string
}

func NewIcinga(conf IcingaConf) Icinga {
	baseUrl := fmt.Sprintf("%s://%s:%d/v1", conf.Connection.Proto, conf.Connection.Server, conf.Connection.Port)
	i := Icinga{
		baseUrl: baseUrl,
		user:    conf.Connection.User,
		pass:    conf.Connection.Pass,
	}
	return i
}

func icingaDefaults() map[string]bool {
	defaults := make(map[string]bool)
	defaults["ok"] = false
	defaults["unknown"] = false
	defaults["warn"] = false
	defaults["critical"] = false
	defaults["scheduled_downtime"] = false
	defaults["acknowledged"] = false
	defaults["failed"] = true
	return defaults
}

func (i Icinga) Values() []string {
	var out []string
	for key, _ := range icingaDefaults() {
		out = append(out, key)
	}
	return out
}

func (i Icinga) Analyze(svc SvcResult) (Status, error) {
	var ok, unknown, warn, critical, scheduledDowntime, acknowledged, failed bool
	var exists bool
	if ok, exists = svc.Vals["ok"]; !exists {
		return StatusUnknown, errors.New("Value 'ok' does not exist")
	} else if unknown, exists = svc.Vals["unknown"]; !exists {
		return StatusUnknown, errors.New("Value 'unknown' does not exist")
	} else if warn, exists = svc.Vals["warn"]; !exists {
		return StatusUnknown, errors.New("Value 'warn' does not exist")
	} else if critical, exists = svc.Vals["critical"]; !exists {
		return StatusUnknown, errors.New("Value 'critical' does not exist")
	} else if scheduledDowntime, exists = svc.Vals["scheduled_downtime"]; !exists {
		return StatusUnknown, errors.New("Value 'scheduled_downtime' does not exist")
	} else if acknowledged, exists = svc.Vals["acknowledged"]; !exists {
		return StatusUnknown, errors.New("Value 'acknowledged' does not exist")
	} else if failed, exists = svc.Vals["failed"]; !exists {
		return StatusUnknown, errors.New("Value 'failed' does not exist")
	}

	if failed || unknown {
		return StatusUnknown, nil
	} else if !scheduledDowntime && critical {
		return StatusNOK, nil
	}
	if ok || warn || acknowledged {
		return StatusOK, nil
	}
	return StatusOK, nil
}

func (i Icinga) Status(s Service) (result SvcResult, err error) {
	result = SvcResult{
		At:   time.Now(),
		Msg:  "",
		Vals: icingaDefaults(),
	}
	err = nil

	// proper encoding for the host string
	hostUrl := &url.URL{Path: s.Host}
	host := hostUrl.String()
	// proper encoding for the service string
	serviceUrl := &url.URL{Path: s.Service}
	service := serviceUrl.String()
	// build url
	url := fmt.Sprintf("%s/objects/services?service=%s!%s", i.baseUrl, host, service)
	// query api
	// TODO: read rootca from file
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(i.user, i.pass)
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.New("HTTP error " + resp.Status)
		return
	}
	// parse response body
	var response serviceStatusResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}
	result, err = response.status()
	return
}

// type serviceStatusResult describes the results returned by the icinga
// api when a service status is requested.
type serviceStatusResponse struct {
	Results []struct {
		Attrs struct {
			Acknowledgement float64 `json:"acknowledgement"`
			//			AcknowledgementExpiry uts     `json:"acknowledgement_expiry"`
			LastCheckResult struct {
				State  float64 `json:"state"`
				Output string  `json:"output"`
			} `json:"last_check_result"`
			LastCheck     Timestamp `json:"last_check"`
			DowntimeDepth float64   `json:"downtime_depth"`
		} `json:"attrs"`
	} `json:"results"`
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

const (
	IcingaStatusOK = iota
	IcingaStatusWarn
	IcingaStatusCritical
	IcingaStatusUnknown
)

func (r serviceStatusResponse) status() (result SvcResult, err error) {
	result = SvcResult{
		At:   time.Now(),
		Msg:  "",
		Vals: icingaDefaults(),
	}

	if len(r.Results) != 1 {
		err = errors.New("Not exactly one Result found in Icinga API response for service")
		return
	}
	attrs := r.Results[0].Attrs

	result.Msg = attrs.LastCheckResult.Output
	result.At = attrs.LastCheck.Time
	result.Vals["failed"] = false
	if attrs.Acknowledgement > 0.0 {
		result.Vals["acknowledged"] = true
	}
	if attrs.DowntimeDepth > 0.0 {
		result.Vals["scheduledDowntime"] = true
	}
	switch attrs.LastCheckResult.State {
	case IcingaStatusOK:
		result.Vals["ok"] = true
	case IcingaStatusWarn:
		result.Vals["warn"] = true
	case IcingaStatusCritical:
		result.Vals["critical"] = true
	case IcingaStatusUnknown:
		result.Vals["unknown"] = true
	default:
		result.Vals["failed"] = true
		err = errors.New("Icinga status unknown")
	}
	return
}
