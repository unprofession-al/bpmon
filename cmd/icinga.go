package cmd

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

const (
	IcingaFlagOK                = "ok"
	IcingaFlagUnknown           = "unknown"
	IcingaFlagWarn              = "warn"
	IcingaFlagCritical          = "critical"
	IcingaFlagScheduledDowntime = "scheduled_downtime"
	IcingaFlagAcknowledged      = "acknowledged"
	IcingaFlagFailed            = "failed"
)

type IcingaConf struct {
	Connection struct {
		Server string
		Port   int
		Pass   string
		User   string
		Proto  string
	}
	Rules []Rule
}

type Rule struct {
	Order      int
	Must       []string
	MustNot    []string
	Then       string
	thenStatus Status
}

type Icinga struct {
	baseUrl string
	user    string
	pass    string
	rules   map[int]Rule
}

func NewIcinga(conf IcingaConf) (Icinga, error) {
	baseUrl := fmt.Sprintf("%s://%s:%d/v1", conf.Connection.Proto, conf.Connection.Server, conf.Connection.Port)

	rules := map[int]Rule{
		10: Rule{
			Must:       []string{IcingaFlagFailed},
			MustNot:    []string{},
			thenStatus: StatusUnknown,
		},
		20: Rule{
			Must:       []string{IcingaFlagUnknown},
			MustNot:    []string{},
			thenStatus: StatusUnknown,
		},
		30: Rule{
			Must:       []string{IcingaFlagCritical},
			MustNot:    []string{IcingaFlagScheduledDowntime},
			thenStatus: StatusNOK,
		},
		9999: Rule{
			Must:       []string{},
			MustNot:    []string{},
			thenStatus: StatusOK,
		},
	}

	for _, r := range conf.Rules {
		status, err := StatusFromString(r.Then)
		if err != nil {
			return Icinga{}, errors.New(fmt.Sprintf("'%s' configured in rule with order %d is not a valid status", r.Then, r.Order))
		}
		rule := Rule{
			Must:       r.Must,
			MustNot:    r.MustNot,
			thenStatus: status,
		}
		rules[r.Order] = rule
	}

	i := Icinga{
		baseUrl: baseUrl,
		user:    conf.Connection.User,
		pass:    conf.Connection.Pass,
		rules:   rules,
	}
	return i, nil
}

func icingaDefaults() map[string]bool {
	defaults := make(map[string]bool)
	defaults[IcingaFlagOK] = false
	defaults[IcingaFlagUnknown] = false
	defaults[IcingaFlagWarn] = false
	defaults[IcingaFlagCritical] = false
	defaults[IcingaFlagScheduledDowntime] = false
	defaults[IcingaFlagAcknowledged] = false
	defaults[IcingaFlagFailed] = true
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
	var order []int
	for index := range i.rules {
		order = append(order, index)
	}
	sort.Ints(order)

	for _, index := range order {
		matchMustCond := true
		matchMustNotCond := true
		rule := i.rules[index]

		for _, keyname := range rule.Must {
			if val, ok := svc.Vals[keyname]; ok {
				if !val {
					matchMustCond = false
					break
				}
			} else {
				return StatusUnknown, errors.New(fmt.Sprintf("Key '%s' from rule with order %d does not exist", keyname, index))
			}
		}

		for _, keyname := range rule.MustNot {
			if val, ok := svc.Vals[keyname]; ok {
				if val {
					matchMustNotCond = false
					break
				}
			} else {
				return StatusUnknown, errors.New(fmt.Sprintf("Key '%s' from rule with order %d does not exist", keyname, index))
			}
		}

		if matchMustCond && matchMustNotCond {
			return rule.thenStatus, nil
		}
	}
	return StatusUnknown, errors.New("No rule matched")
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
	result.Vals[IcingaFlagFailed] = false
	if attrs.Acknowledgement > 0.0 {
		result.Vals[IcingaFlagAcknowledged] = true
	}
	if attrs.DowntimeDepth > 0.0 {
		result.Vals[IcingaFlagScheduledDowntime] = true
	}
	switch attrs.LastCheckResult.State {
	case IcingaStatusOK:
		result.Vals[IcingaFlagOK] = true
	case IcingaStatusWarn:
		result.Vals[IcingaFlagWarn] = true
	case IcingaStatusCritical:
		result.Vals[IcingaFlagCritical] = true
	case IcingaStatusUnknown:
		result.Vals[IcingaFlagUnknown] = true
	default:
		result.Vals[IcingaFlagFailed] = true
		err = errors.New("Icinga status unknown")
	}
	return
}
