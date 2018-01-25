package icinga

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/status"
)

// init registers the 'Checker' implementation.
func init() {
	checker.Register("icinga", Setup)
}

// Setup configures the 'Checker' implementation and returns it.
func Setup(conf checker.Conf) (checker.Checker, error) {
	u, err := url.Parse(conf.Connection)
	if err != nil {
		return nil, err
	}
	username := u.User.Username()
	password, _ := u.User.Password()

	baseURL := fmt.Sprintf("%s://%s%s/v1", u.Scheme, u.Host, u.Path)
	fetcher := api{
		baseURL:       baseURL,
		pass:          password,
		user:          username,
		tlsSkipVerify: conf.TLSSkipVerify,
	}

	i := Icinga{
		f: fetcher,
	}

	return i, nil
}

type flag string

const (
	FlagOK                flag = "ok"
	FlagUnknown           flag = "unknown"
	FlagWarn              flag = "warn"
	FlagCritical          flag = "critical"
	FlagScheduledDowntime flag = "scheduled_downtime"
	FlagAcknowledged      flag = "acknowledged"
	FlagFailed            flag = "failed"
)

func (f flag) String() string {
	return string(f)
}

type flags map[flag]bool

var flagDefaults = flags{
	FlagOK:                false,
	FlagUnknown:           false,
	FlagWarn:              false,
	FlagCritical:          false,
	FlagScheduledDowntime: false,
	FlagAcknowledged:      false,
	FlagFailed:            true,
}

func (f flags) ToValues() map[string]bool {
	out := make(map[string]bool)
	for k, v := range flagDefaults {
		out[k.String()] = v
	}
	return out
}

// Icinga holds the 'Checker' implementation. It allows BPMON to fetch the status
// configured via the Icinga2 API
type Icinga struct {
	f fetcher
}

type fetcher interface {
	Fetch(string, string) (Response, error)
}

// DefaultRules implements the 'Checker' interface.
func (i Icinga) DefaultRules() rules.Rules {
	rules := rules.Rules{
		10: rules.Rule{
			Must:    []string{FlagFailed.String()},
			MustNot: []string{},
			Then:    status.StatusUnknown,
		},
		20: rules.Rule{
			Must:    []string{FlagUnknown.String()},
			MustNot: []string{},
			Then:    status.StatusUnknown,
		},
		30: rules.Rule{
			Must:    []string{FlagCritical.String()},
			MustNot: []string{FlagScheduledDowntime.String()},
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

// Values implements the 'Checker' interface.
func (i Icinga) Values() []string {
	var out []string
	for key := range flagDefaults {
		out = append(out, key.String())
	}
	return out
}

// Status implements the 'Checker' interface.
func (i Icinga) Status(host string, service string) checker.Result {
	r := checker.Result{
		Timestamp: time.Now(),
		Values:    flagDefaults.ToValues(),
	}

	response, err := i.f.Fetch(host, service)
	if err != nil {
		r.Error = err
		return r
	}

	r.Timestamp, r.Message, r.Values, r.Error = response.status()
	return r
}

func (r Response) status() (at time.Time, msg string, vals map[string]bool, err error) {
	at = time.Now()
	msg = ""
	vals = flagDefaults.ToValues()

	if len(r.Results) != 1 {
		err = errors.New("Not exactly one Result found in Icinga API response for service")
		return
	}
	attrs := r.Results[0].Attrs

	msg = attrs.LastCheckResult.Output
	at = time.Time(attrs.LastCheck)
	vals[FlagFailed.String()] = false
	if attrs.Acknowledgement > 0.0 {
		vals[FlagAcknowledged.String()] = true
	}
	if attrs.DowntimeDepth > 0.0 {
		vals[FlagScheduledDowntime.String()] = true
	}
	switch attrs.LastCheckResult.State {
	case statusOK:
		vals[FlagOK.String()] = true
	case statusWarn:
		vals[FlagWarn.String()] = true
	case statusCritical:
		vals[FlagCritical.String()] = true
	case statusUnknown:
		vals[FlagUnknown.String()] = true
	default:
		vals[FlagFailed.String()] = true
		err = errors.New("Icinga status unknown")
	}
	return
}

type api struct {
	baseURL       string
	user          string
	pass          string
	tlsSkipVerify bool
}

func (a api) Fetch(host, service string) (Response, error) {
	var response Response
	var body []byte

	// proper encoding for the host string
	hostURL := &url.URL{Path: host}
	host = hostURL.String()
	// proper encoding for the service string
	serviceURL := &url.URL{Path: service}
	service = serviceURL.String()
	// build url
	url := fmt.Sprintf("%s/objects/services?service=%s!%s", a.baseURL, host, service)
	// query api
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: a.tlsSkipVerify},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}
	req.SetBasicAuth(a.user, a.pass)
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
