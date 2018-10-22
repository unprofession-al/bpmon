package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/unprofession-al/bpmon/availabilities"
	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/dashboard"
	"github.com/unprofession-al/bpmon/health"
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

func NewFromFile(path string, inject bool) (c Config, raw []byte, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return c, data, fmt.Errorf("Error while reading configuration file '%s': %s", path, err.Error())
	}
	return New(data, inject)
}

func New(data []byte, inject bool) (c Config, raw []byte, err error) {
	if inject {
		key := "injected_defaults_" + time.Now().Format("20060102150405")
		raw, err = injectDefaults(data, key)
		if err != nil {
			return c, raw, fmt.Errorf("Error while injection defaults: %s", err.Error())
		}
	} else {
		raw = data
	}

	err = yaml.Unmarshal(raw, &c)
	if err != nil {
		return c, raw, fmt.Errorf("Error while unmarshalling configuration from yaml: %s", err.Error())
	}
	return c, raw, nil
}

func ExampleYAML(inject bool) []byte {
	section := ConfigDefaultSection
	defaultData := Config{section: defaultConfigSection()}
	example, _ := yaml.Marshal(defaultData)
	if inject {
		example = injectComments(example, section)
	}
	return example
}

type ConfigSection struct {
	// global_recipients will be added to the repicients list af all BP
	GlobalRecipients []string `yaml:"global_recipients"`

	// health ... TODO
	Health health.Config `yaml:"health"`

	// First BPMON needs to have access to your Icinga2 API. Learn more on by reading
	// https://docs.icinga.com/icinga2/latest/doc/module/icinga2/chapter/icinga2-api.
	Checker checker.Config `yaml:"checker"`

	// Also the connection to the InfluxDB is required in order to persist the
	// state for reporting and such
	Store store.Config `yaml:"store"`

	// Define your office hours et al. according to your service level
	// agreements (SLA). You can later reference them in your BP definitions.
	Availabilities availabilities.AvailabilitiesConfig `yaml:"availabilities"`

	// Extend the default rules; in that case: Do not run the alarming command
	// if a critical service is aready aknowledged to avoid alarm spamming.
	Rules rules.Rules `yaml:"rules"`

	// dashboard configures the dashboard subcommand.
	Dashboard dashboard.Config `yaml:"dashboard"`

	Env EnvConfig `yaml:"env"`
}

func defaultConfigSection() ConfigSection {
	return ConfigSection{
		Health:    health.Defaults(),
		Checker:   checker.Defaults(),
		Store:     store.Defaults(),
		Dashboard: dashboard.Defaults(),
		Env:       EnvDefaults(),
	}
}

func (s ConfigSection) Validate(name string) (out []string, err error) {
	var errs []string

	errs = fmtErrors(s.Health.Validate())
	out = append(out, errs...)

	errs = fmtErrors(s.Checker.Validate())
	out = append(out, errs...)

	errs = fmtErrors(s.Store.Validate())
	out = append(out, errs...)

	errs = fmtErrors(s.Dashboard.Validate())
	out = append(out, errs...)

	errs = fmtErrors(s.Env.Validate())
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
