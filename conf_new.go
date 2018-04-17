package bpmon

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/store"
	yaml "gopkg.in/yaml.v2"
)

type Conf map[string]ConfSection

func (c Conf) Validate() (out []string, err error) {
	var errs []string
	for n, s := range c {
		errs = fmtErrors(s.Validate(n))
		out = append(out, errs...)
	}
	if len(errs) > 0 {
		err = errors.New("Configuration has errors")
	}
	return
}

type ConfSection struct {
	Health         HealthConfig       `yaml:"health"`
	Checker        checker.Config     `yaml:"checker"`
	Store          store.Config       `yaml:"store"`
	Availabilities AvailabilitiesConf `yaml:"availabilities"`
}

func Load(path string) (Conf, error) {
	c := ConfDefaults()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("Error while reading configuration file '%s': %s", path, err.Error())
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, fmt.Errorf("Error while unmarshalling configuration from yaml: %s", err.Error())
	}
	return c, nil
}

func ConfDefaults() Conf {
	c := make(Conf)
	c["default"] = ConfSection{
		Health:  HealthConfigDefaults(),
		Checker: checker.ConfigDefaults(),
		Store:   store.ConfigDefaults(),
	}
	return c
}

func (s ConfSection) Validate(name string) (out []string, err error) {
	var errs []string

	errs = fmtErrors(s.Health.Validate())
	out = append(out, errs...)

	errs = fmtErrors(s.Checker.Validate())
	out = append(out, errs...)

	errs = fmtErrors(s.Store.Validate())
	out = append(out, errs...)

	_, aErr := s.Availabilities.Parse()
	if aErr != nil {
		out = append(out, fmt.Sprintf("Error while parsing availabilities: %s", aErr.Error()))
	}

	if len(errs) > 0 {
		err = fmt.Errorf("Configuration Section '%s' has errors", name)
	}
	return
}

func fmtErrors(errs []string, err error) []string {
	out := []string{}
	if err != nil {
		for _, msg := range errs {
			out = append(out, fmt.Sprintf("%s: %s", err, msg))
		}
	}
	return out
}

func (c Conf) Section(name string) ConfSection {
	if section, ok := c[name]; ok {
		return section
	}
	return ConfSection{}
}
