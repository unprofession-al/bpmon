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

func (i Icinga) ServiceStatus(s Service) (ok bool, inDowntime bool, output string, err error) {
	err = nil
	inDowntime = false
	output = ""
	ok = true

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
	var results serviceStatusResults
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &results)
	if err != nil {
		return
	}
	ok, inDowntime, output, err = results.status()
	return
}

// type serviceStatusResult describes the results returned by the icinga
// api when a service status is requested.
type serviceStatusResults struct {
	Results []struct {
		Attrs struct {
			Acknowledgement float64 `json:"acknowledgement"`
			//			AcknowledgementExpiry uts     `json:"acknowledgement_expiry"`
			LastCheckResult struct {
				State  float64 `json:"state"`
				Output string  `json:"output"`
			} `json:"last_check_result"`
			LastCheck      Timestamp `json:"last_check"`
			LastInDowntime bool      `json:"last_in_downtime"`
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

func (r serviceStatusResults) status() (ok bool, inDowntime bool, output string, err error) {
	err = nil
	inDowntime = false
	ok = true
	output = ""

	if len(r.Results) != 1 {
		err = errors.New("not exactly one result found")
		return
	}
	output = r.Results[0].Attrs.LastCheckResult.Output
	if r.Results[0].Attrs.LastInDowntime {
		inDowntime = true
	}
	if r.Results[0].Attrs.LastCheckResult.State == IcingaStatusCritical {
		ok = false
	}
	return
}
