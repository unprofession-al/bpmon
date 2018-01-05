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
	fetcher := API{
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
	FlagOK                = "ok"
	FlagUnknown           = "unknown"
	FlagWarn              = "warn"
	FlagCritical          = "critical"
	FlagScheduledDowntime = "scheduled_downtime"
	FlagAcknowledged      = "acknowledged"
	FlagFailed            = "failed"
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
	Fetch(string, string) (Response, error)
}

func (i Icinga) DefaultRules() rules.Rules {
	rules := rules.Rules{
		10: rules.Rule{
			Must:    []string{FlagFailed},
			MustNot: []string{},
			Then:    status.StatusUnknown,
		},
		20: rules.Rule{
			Must:    []string{FlagUnknown},
			MustNot: []string{},
			Then:    status.StatusUnknown,
		},
		30: rules.Rule{
			Must:    []string{FlagCritical},
			MustNot: []string{FlagScheduledDowntime},
			Then:    status.StatusNOK,
		},
		9999: rules.Rule{
			Must:    []string{},
			MustNot: []string{},
			Then:    status.StatusOK,
		},
	}
	return rules
}

func icingaDefaultFlags() map[string]bool {
	defaults := make(map[string]bool)
	defaults[FlagOK] = false
	defaults[FlagUnknown] = false
	defaults[FlagWarn] = false
	defaults[FlagCritical] = false
	defaults[FlagScheduledDowntime] = false
	defaults[FlagAcknowledged] = false
	defaults[FlagFailed] = true
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

// Response describes the results returned by the icinga
// api when a service status is requested.
type Response struct {
	Results []StatusResult `json:"results"`
}

type StatusResult struct {
	Attrs StatusAttrs `json:"attrs"`
	Name  string      `json:"name"`
}

type StatusAttrs struct {
	Acknowledgement float64         `json:"acknowledgement"`
	LastCheckResult LastCheckResult `json:"last_check_result"`
	LastCheck       Timestamp       `json:"last_check"`
	DowntimeDepth   float64         `json:"downtime_depth"`
}

type LastCheckResult struct {
	State  float64 `json:"state"`
	Output string  `json:"output"`
}

type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	tsString, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}

	ts := time.Unix(int64(tsString), 0)
	*t = Timestamp(ts)

	return nil
}

const (
	StatusOK = iota
	StatusWarn
	StatusCritical
	StatusUnknown
)

func (r Response) status() (at time.Time, msg string, vals map[string]bool, err error) {
	at = time.Now()
	msg = ""
	vals = icingaDefaultFlags()

	if len(r.Results) != 1 {
		err = errors.New("Not exactly one Result found in Icinga API response for service")
		return
	}
	attrs := r.Results[0].Attrs

	msg = attrs.LastCheckResult.Output
	at = time.Time(attrs.LastCheck)
	vals[FlagFailed] = false
	if attrs.Acknowledgement > 0.0 {
		vals[FlagAcknowledged] = true
	}
	if attrs.DowntimeDepth > 0.0 {
		vals[FlagScheduledDowntime] = true
	}
	switch attrs.LastCheckResult.State {
	case StatusOK:
		vals[FlagOK] = true
	case StatusWarn:
		vals[FlagWarn] = true
	case StatusCritical:
		vals[FlagCritical] = true
	case StatusUnknown:
		vals[FlagUnknown] = true
	default:
		vals[FlagFailed] = true
		err = errors.New("Icinga status unknown")
	}
	return
}

type API struct {
	baseURL string
	user    string
	pass    string
}

func (i API) Fetch(host, service string) (Response, error) {
	var response Response
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
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	return response, err
}
