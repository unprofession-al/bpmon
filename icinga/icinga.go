package icinga

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

	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
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
	Server string
	Port   int
	Pass   string
	User   string
	Proto  string
}

type Icinga struct {
	fecher IcingaFetcher
	rules  rules.Rules
}

type IcingaFetcher interface {
	Fetch(string, string) ([]byte, error)
}

func NewIcinga(conf IcingaConf, additionalRules rules.Rules) (Icinga, error) {
	r := icingaDefaultRules()
	for order, ar := range additionalRules {
		s, err := status.FromString(ar.Then)
		if err != nil {
			return Icinga{}, errors.New(fmt.Sprintf("'%s' configured in rule with order '%d' is not a valid status", ar.Then, order))
		}
		rule := rules.Rule{
			Must:       ar.Must,
			MustNot:    ar.MustNot,
			ThenStatus: s,
		}
		r[order] = rule
	}

	baseUrl := fmt.Sprintf("%s://%s:%d/v1", conf.Proto, conf.Server, conf.Port)
	fetcher := IcingaAPI{
		baseUrl: baseUrl,
		pass:    conf.Pass,
		user:    conf.User,
	}

	i := Icinga{
		fecher: fetcher,
		rules:  r,
	}
	return i, nil
}

func (i Icinga) Rules() rules.Rules {
	return i.rules
}

func icingaDefaultRules() rules.Rules {
	rules := rules.Rules{
		10: rules.Rule{
			Must:       []string{IcingaFlagFailed},
			MustNot:    []string{},
			ThenStatus: status.Unknown,
		},
		20: rules.Rule{
			Must:       []string{IcingaFlagUnknown},
			MustNot:    []string{},
			ThenStatus: status.Unknown,
		},
		30: rules.Rule{
			Must:       []string{IcingaFlagCritical},
			MustNot:    []string{IcingaFlagScheduledDowntime},
			ThenStatus: status.Nok,
		},
		9999: rules.Rule{
			Must:       []string{},
			MustNot:    []string{},
			ThenStatus: status.Ok,
		},
	}
	return rules
}

func icingaDefaultFlags() map[string]bool {
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
	for key, _ := range icingaDefaultFlags() {
		out = append(out, key)
	}
	return out
}

func (i Icinga) Status(host string, service string) (at time.Time, msg string, vals map[string]bool, err error) {
	at = time.Now()
	msg = ""
	vals = icingaDefaultFlags()

	body, err := i.fecher.Fetch(host, service)

	var response serviceStatusResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}
	at, msg, vals, err = response.status()
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

func (r serviceStatusResponse) status() (at time.Time, msg string, vals map[string]bool, err error) {
	at = time.Now()
	msg = ""
	vals = icingaDefaultFlags()

	if len(r.Results) != 1 {
		err = errors.New("Not exactly one Result found in Icinga API response for service")
		return
	}
	attrs := r.Results[0].Attrs

	msg = attrs.LastCheckResult.Output
	at = attrs.LastCheck.Time
	vals[IcingaFlagFailed] = false
	if attrs.Acknowledgement > 0.0 {
		vals[IcingaFlagAcknowledged] = true
	}
	if attrs.DowntimeDepth > 0.0 {
		vals[IcingaFlagScheduledDowntime] = true
	}
	switch attrs.LastCheckResult.State {
	case IcingaStatusOK:
		vals[IcingaFlagOK] = true
	case IcingaStatusWarn:
		vals[IcingaFlagWarn] = true
	case IcingaStatusCritical:
		vals[IcingaFlagCritical] = true
	case IcingaStatusUnknown:
		vals[IcingaFlagUnknown] = true
	default:
		vals[IcingaFlagFailed] = true
		err = errors.New("Icinga status unknown")
	}
	return
}

type IcingaAPI struct {
	baseUrl string
	user    string
	pass    string
}

func (i IcingaAPI) Fetch(host, service string) ([]byte, error) {
	var body []byte
	// proper encoding for the host string
	hostUrl := &url.URL{Path: host}
	host = hostUrl.String()
	// proper encoding for the service string
	serviceUrl := &url.URL{Path: service}
	service = serviceUrl.String()
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
		return body, err
	}
	req.SetBasicAuth(i.user, i.pass)
	resp, err := client.Do(req)
	if err != nil {
		return body, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.New("HTTP error " + resp.Status)
		return body, err
	}
	// parse response body
	body, err = ioutil.ReadAll(resp.Body)
	return body, err
}
