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

// Icinga holds the 'Checker' implementation. It allows BPMON to fetch the status
// configured via the Icinga2 API
type Icinga struct {
	f fetcher
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

// Health implements the 'Checker' interface.
func (i Icinga) Health() (string, error) {
	return i.f.Health()
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

type fetcher interface {
	Fetch(string, string) (Response, error)
	Health() (string, error)
}

type api struct {
	baseURL       string
	user          string
	pass          string
	tlsSkipVerify bool
}

func (a api) Fetch(host, service string) (Response, error) {
	var response Response

	// proper encoding for the host string
	hostURL := &url.URL{Path: host}
	host = hostURL.String()
	// proper encoding for the service string
	serviceURL := &url.URL{Path: service}
	service = serviceURL.String()
	// build url
	url := fmt.Sprintf("%s/objects/services?service=%s!%s", a.baseURL, host, service)
	body, err := a.get(url)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	return response, err
}

func (a api) Health() (string, error) {
	url := fmt.Sprintf("%s/status", a.baseURL)
	body, err := a.get(url)
	return string(body), err
}

func (a api) get(url string) ([]byte, error) {
	var body []byte

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: a.tlsSkipVerify},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return body, err
	}
	req.SetBasicAuth(a.user, a.pass)
	resp, err := client.Do(req)
	if err != nil {
		return body, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.New("HTTP error " + resp.Status)
		return body, err
	}

	return ioutil.ReadAll(resp.Body)
}
