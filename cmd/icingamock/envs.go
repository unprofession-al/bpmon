package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/internal/checker/icinga"
	yaml "gopkg.in/yaml.v2"
)

type Environments map[string]*Hosts

func (e Environments) ToIcinga(envN string, t icinga.Timestamp) (icinga.Response, error) {
	response := icinga.Response{}

	env, ok := e[envN]
	if !ok {
		return response, fmt.Errorf("Environment %s unknown", envN)
	}
	for hostname, services := range *env {
		for servicename, service := range *services {
			result := icinga.Result{
				Attrs: icinga.Attrs{
					Acknowledgement: Btof(service.Acknowledgement),
					DowntimeDepth:   Btof(service.Downtime),
					LastCheck:       t,
					LastCheckResult: icinga.LastCheckResult{
						State:  float64(service.CheckState),
						Output: service.CheckOutput,
					},
				},
				Name: fmt.Sprintf("%s!%s", hostname, servicename),
			}
			response.Results = append(response.Results, result)
		}
	}
	return response, nil
}

func (e Environments) SingleToIcinga(envN, hostN, serviceN string, t icinga.Timestamp) (icinga.Response, error) {
	response := icinga.Response{}

	env, ok := e[envN]
	if !ok {
		return response, fmt.Errorf("Environment %s unknown", envN)
	}
	for hostname, services := range *env {
		if hostname == hostN {
			for servicename, service := range *services {
				if servicename == serviceN {
					result := icinga.Result{
						Attrs: icinga.Attrs{
							Acknowledgement: Btof(service.Acknowledgement),
							DowntimeDepth:   Btof(service.Downtime),
							LastCheck:       t,
							LastCheckResult: icinga.LastCheckResult{
								State:  float64(service.CheckState),
								Output: service.CheckOutput,
							},
						},
						Name: fmt.Sprintf("%s!%s", hostname, servicename),
					}
					response.Results = append(response.Results, result)
				}
			}
		}
	}
	return response, nil
}

func (e Environments) Get(name string) (*Hosts, error) {
	for n, env := range e {
		if n == name {
			return env, nil
		}
	}
	return &Hosts{}, fmt.Errorf("Environment %s not found", name)
}

func (e Environments) List() []string {
	var out []string
	for n := range e {
		out = append(out, n)
	}
	return out
}

func Btof(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func LoadEnvs(path, pattern string) (*Environments, error) {
	e := &Environments{}
	if path == "" {
		return e, nil
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return e, fmt.Errorf("Error while reading environment configuration files from '%s': %s", path, err.Error())
	}
	for _, f := range files {
		match, err := filepath.Match(pattern, f.Name())
		if err != nil {
			return e, fmt.Errorf("Error while matching file pattern '%s' in '%s': %s", pattern, path, err.Error())
		}
		if !match {
			continue
		}
		file, err := ioutil.ReadFile(path + "/" + f.Name())
		if err != nil {
			return e, fmt.Errorf("Error while reading file %s/%s: %s", path, f.Name(), err.Error())
		}
		hosts, err := parseEnv(file)
		if err != nil {
			return e, fmt.Errorf("Error while reading environment %s/%s: %s", path, f.Name(), err.Error())
		}
		envName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		(*e)[envName] = hosts
	}
	return e, nil
}

func parseEnv(data []byte) (*Hosts, error) {
	var env *Hosts
	err := yaml.Unmarshal(data, &env)
	if err != nil {
		return env, fmt.Errorf("Error while parsing: %s", err.Error())
	}
	return env, nil
}

func LoadEnvFromBP(path, pattern string) (*Hosts, error) {
	e := &Hosts{}
	if path == "" {
		return e, nil
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return e, fmt.Errorf("Error while reading business configuration files from '%s': %s", path, err.Error())
	}
	for _, f := range files {
		match, err := filepath.Match(pattern, f.Name())
		if err != nil {
			return e, fmt.Errorf("Error while matching file pattern '%s' in '%s': %s", pattern, path, err.Error())
		}
		if !match {
			continue
		}
		file, err := ioutil.ReadFile(path + "/" + f.Name())
		if err != nil {
			return e, fmt.Errorf("Error while reading file %s/%s: %s", path, f.Name(), err.Error())
		}
		e, err = parseBP(file, e)
		if err != nil {
			return e, fmt.Errorf("Error while reading business process %s/%s: %s", path, f.Name(), err.Error())
		}
	}
	return e, nil
}

func parseBP(bpconf []byte, env *Hosts) (*Hosts, error) {
	bp := bpmon.BP{}
	err := yaml.Unmarshal(bpconf, &bp)
	if err != nil {
		return env, fmt.Errorf("Error while parsing: %s", err.Error())
	}

	for _, kpi := range bp.Kpis {
		for _, svc := range kpi.Services {
			_, ok := (*env)[svc.Host]
			if !ok {
				(*env)[svc.Host] = &Services{}
			}
			_, ok = (*(*env)[svc.Host])[svc.Service]
			if !ok {
				(*(*env)[svc.Host])[svc.Service] = &Service{CheckOutput: "ok"}
			}

		}
	}

	return env, nil
}
