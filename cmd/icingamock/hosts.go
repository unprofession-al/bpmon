package main

import "fmt"

type Hosts map[string]*Services

func (h *Hosts) Get(name string) (*Services, error) {
	for n, host := range *h {
		if n == name {
			return host, nil
		}
	}
	return &Services{}, fmt.Errorf("Host %s not found", name)
}

func (h Hosts) List() []string {
	var out []string
	for n := range h {
		out = append(out, n)
	}
	return out
}
