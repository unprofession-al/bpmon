package cmd

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type IcingaConf struct {
	Connection struct {
		Server string
		Port   int
		Pass   string
		User   string
		Proto  string
	}
	Timeout int
}

type Icinga struct {
	baseUrl string
	user    string
	pass    string
	timeout int
}

func NewIcinga(conf IcingaConf) Icinga {
	baseUrl := fmt.Sprintf("%s://%s:%d/v1", conf.Connection.Proto, conf.Connection.Server, conf.Connection.Port)
	i := Icinga{
		baseUrl: baseUrl,
		user:    conf.Connection.User,
		pass:    conf.Connection.Pass,
		timeout: conf.Timeout,
	}
	return i
}

func (i Icinga) ServiceStatus(s Service) (bool, error) {
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
		return true, err
	}
	req.SetBasicAuth(i.user, i.pass)
	resp, err := client.Do(req)
	if err != nil {
		return true, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return true, errors.New("HTTP error " + resp.Status)
	}
	// parse response body
	var results serviceStatusResults
	body, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &results); err != nil {
		return true, err
	}
	return results.status()
}

// type serviceStatusResult describes the results returned by the icinga
// api when a service status is requested.
type serviceStatusResults struct {
	Results []struct {
		Attrs struct {
			LastCheckResult struct {
				State float64 `json:"state"`
			} `json:"last_check_result"`
			LastInDowntime bool `json:"last_in_downtime"`
		} `json:"attrs"`
	} `json:"results"`
}

const (
	IcingaStatusOK = iota
	IcingaStatusWarn
	IcingaStatusCritical
	IcingaStatusUnknown
)

func (r serviceStatusResults) status() (bool, error) {
	if len(r.Results) != 1 {
		return true, errors.New("not exactly one result found")
	}
	// only fail if critical and not in scheduled downtime
	// TODO: consider to tread Downtime as a separate state
	// status Downtime could be a tag in the time series
	if r.Results[0].Attrs.LastCheckResult.State == IcingaStatusCritical && !r.Results[0].Attrs.LastInDowntime {
		return false, nil
	}
	return true, nil
}
