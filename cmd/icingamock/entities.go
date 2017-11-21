package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Environments map[string]Hosts

func (e Environments) ToIcinga(envN string, t Timestamp) (IcingaStatusResponse, error) {
	response := IcingaStatusResponse{}

	env, ok := e[envN]
	if !ok {
		return response, fmt.Errorf("Environment %s unknown", envN)
	}
	for hostname, services := range env {
		for servicename, service := range services {
			result := IcingaStatusResult{
				Attrs: IcingaStatusAttrs{
					Acknowledgement: Btof(service.Acknowledgement),
					DowntimeDepth:   Btof(service.Downtime),
					LastCheck:       t,
					LastCheckResult: IcingaStatusLastCheckResult{
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

func (e Environments) SingleToIcinga(envN, hostN, serviceN string, t Timestamp) (IcingaStatusResponse, error) {
	response := IcingaStatusResponse{}

	env, ok := e[envN]
	if !ok {
		return response, fmt.Errorf("Environment %s unknown", envN)
	}
	for hostname, services := range env {
		if hostname == hostN {
			for servicename, service := range services {
				if servicename == serviceN {
					result := IcingaStatusResult{
						Attrs: IcingaStatusAttrs{
							Acknowledgement: Btof(service.Acknowledgement),
							DowntimeDepth:   Btof(service.Downtime),
							LastCheck:       t,
							LastCheckResult: IcingaStatusLastCheckResult{
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

type Hosts map[string]map[string]Service

type Service struct {
	CheckState      int    `yaml:"check_state"`
	CheckOutput     string `yaml:"check_output"`
	Acknowledgement bool   `yaml:"acknowledgement"`
	Downtime        bool   `yaml:"downtime"`
}

func Btof(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func LoadEnvs(path, pattern string) (Environments, error) {
	e := Environments{}
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
		hosts, err := parseEnv(file)
		if err != nil {
			return e, fmt.Errorf("Error while reading environment %s/%s: %s", path, f.Name(), err.Error())
		}
		envName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		e[envName] = hosts
	}
	return e, nil
}

func parseEnv(data []byte) (Hosts, error) {
	var hosts Hosts
	err := yaml.Unmarshal(data, &hosts)
	if err != nil {
		return hosts, fmt.Errorf("Error while parsing: %s", err.Error())
	}
	//for hostname, services := range hosts {
	//	for servicename, service := range services {
	//		service.HostName = hostname
	//		service.ServiceName = servicename
	//	}
	//}
	return hosts, nil
}
