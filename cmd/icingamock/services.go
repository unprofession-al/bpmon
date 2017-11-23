package main

import "fmt"

type Services map[string]*Service

func (s *Services) Get(name string) (*Service, error) {
	for n, service := range *s {
		if n == name {
			return service, nil
		}
	}
	return &Service{}, fmt.Errorf("Service %s not found", name)
}

func (s Services) List() []string {
	var out []string
	for n, _ := range s {
		out = append(out, n)
	}
	return out
}

type Service struct {
	CheckState      int    `yaml:"check_state" json:"check_state"`
	CheckOutput     string `yaml:"check_output" json:"check_output"`
	Acknowledgement bool   `yaml:"acknowledgement" json:"acknowledgement"`
	Downtime        bool   `yaml:"downtime" json:"downtime"`
}
