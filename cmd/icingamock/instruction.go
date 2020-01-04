package main

import (
	"fmt"
)

type Instruction struct {
	Env     string                 `json:"env"`
	Host    string                 `json:"host"`
	Service string                 `json:"service"`
	Attrs   map[string]interface{} `json:"attrs"`
}

func (i Instruction) Apply() error {
	env, err := envs.Get(i.Env)
	if err != nil {
		return err
	}
	host, err := env.Get(i.Host)
	if err != nil {
		return err
	}
	svc, err := host.Get(i.Service)
	if err != nil {
		return err
	}

	for name, value := range i.Attrs {
		switch name {
		case "state":
			state := int(value.(float64))
			svc.CheckState = state
		case "downtime":
			downtime := value.(bool)
			svc.Downtime = downtime
		case "acknowledgement":
			ack := value.(bool)
			svc.Acknowledgement = ack
		default:
			return fmt.Errorf("unknown instruction %s", name)
		}
	}
	return nil
}
