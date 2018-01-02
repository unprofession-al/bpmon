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

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
)

func init() {
	checker.Register("icinga", Setup)
}

func Setup(conf checker.Conf) (checker.Checker, error) {
	u, err := url.Parse(conf.Connection)
	if err != nil {
		panic(err)
	}
	username := u.User.Username()
	password, _ := u.User.Password()

	baseURL := fmt.Sprintf("%s://%s%s/v1", u.Scheme, u.Host, u.Path)
	fetcher := IcingaAPI{
		baseURL: baseURL,
		pass:    password,
		user:    username,
	}

	i := Icinga{
		fecher: fetcher,
	}

	return i, nil
}

const (
	IcingaFlagOK                = "ok"
	IcingaFlagUnknown           = "unknown"
	IcingaFlagWarn              = "warn"
	IcingaFlagCritical          = "critical"
	IcingaFlagScheduledDowntime = "scheduled_downtime"
	IcingaFlagAcknowledged      = "acknowledged"
	IcingaFlagFailed            = "failed"
)

type Conf struct {
	Server string `yaml:"server"`
	Path   string `yaml:"path"`
	Port   int    `yaml:"port"`
	Pass   string `yaml:"pass"`
	User   string `yaml:"user"`
	Proto  string `yaml:"proto"`
}

type Icinga struct {
	fecher Fetcher
}

type Fetcher interface {
	Fetch(string, string) (StatusResponse, error)
}

func (i Icinga) DefaultRules() rules.Rules {
	rules := rules.Rules{
		10: rules.Rule{
			Must:    []string{IcingaFlagFailed},
			MustNot: []string{},
			Then:    status.Unknown,
		},
		20: rules.Rule{
			Must:    []string{IcingaFlagUnknown},
			MustNot: []string{},
			Then:    status.Unknown,
		},
		30: rules.Rule{
			Must:    []string{IcingaFlagCritical},
			MustNot: []string{IcingaFlagScheduledDowntime},
			Then:    status.NOK,
		},
		9999: rules.Rule{
			Must:    []string{},
			MustNot: []string{},
			Then:    status.OK,
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
	for key := range icingaDefaultFlags() {
		out = append(out, key)
	}
	return out
}

func (i Icinga) Status(host string, service string) checker.Result {
	r := checker.Result{
		Timestamp: time.Now(),
		Values:    icingaDefaultFlags(),
	}

	response, err := i.fecher.Fetch(host, service)
	if err != nil {
		r.Error = err
		return r
	}

	r.Timestamp, r.Message, r.Values, r.Error = response.status()
	return r
}

// IcingaStatusResult describes the results returned by the icinga
// api when a service status is requested.
type StatusResponse struct {
	Results []IcingaStatusResult `json:"results"`
}

type IcingaStatusResult struct {
	Attrs IcingaStatusAttrs `json:"attrs"`
	Name  string            `json:"name"`
}

type IcingaStatusAttrs struct {
	Acknowledgement float64                     `json:"acknowledgement"`
	LastCheckResult IcingaStatusLastCheckResult `json:"last_check_result"`
	LastCheck       Timestamp                   `json:"last_check"`
	DowntimeDepth   float64                     `json:"downtime_depth"`
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

const (
	IcingaStatusOK = iota
	IcingaStatusWarn
	IcingaStatusCritical
	IcingaStatusUnknown
)

func (r StatusResponse) status() (at time.Time, msg string, vals map[string]bool, err error) {
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
	baseURL string
	user    string
	pass    string
}

func (i IcingaAPI) Fetch(host, service string) (StatusResponse, error) {
	var response StatusResponse
	var body []byte

	// proper encoding for the host string
	hostURL := &url.URL{Path: host}
	host = hostURL.String()
	// proper encoding for the service string
	serviceURL := &url.URL{Path: service}
	service = serviceURL.String()
	// build url
	url := fmt.Sprintf("%s/objects/services?service=%s!%s", i.baseURL, host, service)
	// query api
	// TODO: read rootca from file
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}
	req.SetBasicAuth(i.user, i.pass)
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.New("HTTP error " + resp.Status)
		return response, err
	}
	// parse response body
	body, err = ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &response)
	return response, err
}
