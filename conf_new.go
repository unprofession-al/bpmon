package bpmon

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/store"
	yaml "gopkg.in/yaml.v2"
)

const ConfigDefaultSection = "default"

type Config map[string]ConfigSection

func (c Config) Validate() (out []string, err error) {
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

type ConfigSection struct {
	Health         HealthConfig       `yaml:"health"`
	Checker        checker.Config     `yaml:"checker"`
	Store          store.Config       `yaml:"store"`
	Availabilities AvailabilitiesConf `yaml:"availabilities"`
	Rules          rules.Rules        `yaml:"rules"`
}

func Load(path string) (Config, error) {
	c := ConfigDefaults()

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

func ConfigDefaults() Config {
	c := make(Config)
	c[ConfigDefaultSection] = ConfigSection{
		Health:  HealthConfigDefaults(),
		Checker: checker.ConfigDefaults(),
		Store:   store.ConfigDefaults(),
	}
	return c
}

func (s ConfigSection) Validate(name string) (out []string, err error) {
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

func (c Config) Section(name string) (ConfigSection, error) {
	if section, ok := c[name]; ok {
		return section, nil
	}
	return ConfigSection{}, fmt.Errorf("Section '%s' not found", name)
}
