package main

import (
	"errors"

	"github.com/unprofession-al/bpmon/configuration"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	configuration.Register("health", unmarshaller)
}

type HealthConfig struct {
	Template        string `yaml:"template"`
	StoreRequired   bool   `yaml:"store_required"`
	CheckerRequired bool   `yaml:"checker_required"`
	Responsible     string `yaml:"responsible"`
	Name            string `yaml:"name"`
	ID              string `yaml:"id"`
}

var defaults = HealthConfig{
	Template:        "{{.}}",
	StoreRequired:   false,
	CheckerRequired: true,
	Responsible:     "",
	Name:            "",
	ID:              "bla",
}

func (hc HealthConfig) Validate() (error, []string) {
	errs := []string{}
	if hc.Template == "" {
		errs = append(errs, "Field 'template' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'health' has errors")
		return err, errs
	}
	return nil, errs
}

func unmarshaller(in interface{}) (configuration.Fragment, error) {
	out := defaults
	data, err := yaml.Marshal(&in)
	if err != nil {
		return defaults, err
	}

	err = yaml.Unmarshal(data, &out)
	if err != nil {
		return defaults, err
	}
	return out, nil
}
