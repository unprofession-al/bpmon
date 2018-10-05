package bpmon

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/unprofession-al/bpmon/availabilities"
	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/health"
	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/store"
	"github.com/unprofession-al/bpmon/trigger"

	"gopkg.in/yaml.v2"
)

type conf struct {
	GlobalRecipient string                              `yaml:"global_recipient"`
	Checker         checker.Config                      `yaml:"checker"`
	Store           store.Config                        `yaml:"store"`
	Availabilities  availabilities.AvailabilitiesConfig `yaml:"availabilities"`
	Trigger         trigger.Config                      `yaml:"trigger"`
	Health          health.Config                       `yaml:"health"`
	Rules           rules.Rules                         `yaml:"rules"`
	Dashboard       config.DashboardConfig              `yaml:"dashboard"`
	Annotate        config.AnnotateConfig               `yaml:"annotate"`
}

func Configure(cfgFile, cfgSection, bpPath, bpPattern string) (conf, BusinessProcesses, error) {
	cfgData, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return conf{}, nil, fmt.Errorf("Error while reading configuration file '%s': %s", cfgFile, err.Error())
	}

	c, err := parseConf(cfgData, cfgSection)
	if err != nil {
		return c, nil, fmt.Errorf("Error while parsing section '%s' configuration file '%s': %s", cfgSection, cfgFile, err.Error())
	}

	if c.Health.Template == "" {
		c.Health.Template = c.Trigger.Template
	}

	a, err := c.Availabilities.Parse()
	if err != nil {
		return c, nil, fmt.Errorf("Error while parsing availabilities from configuration file '%s': %s", cfgFile, err.Error())
	}

	var bps BusinessProcesses
	if bpPath == "" {
		return c, bps, nil
	}
	files, err := ioutil.ReadDir(bpPath)
	if err != nil {
		return c, bps, fmt.Errorf("Error while reading business configuration files from '%s': %s", bpPath, err.Error())
	}
	for _, f := range files {
		match, err := filepath.Match(bpPattern, f.Name())
		if err != nil {
			return c, bps, fmt.Errorf("Error while matching file pattern '%s' in '%s': %s", bpPattern, bpPath, err.Error())
		}
		if !match {
			continue
		}
		file, err := ioutil.ReadFile(bpPath + "/" + f.Name())
		if err != nil {
			return c, bps, fmt.Errorf("Error while reading business process %s/%s: %s", bpPath, f.Name(), err.Error())
		}
		bp, err := parseBP(file, a, c.GlobalRecipient)
		if err != nil {
			return c, bps, fmt.Errorf("Error while parsing business process%s/%s: %s", bpPath, f.Name(), err.Error())
		}
		bps = append(bps, bp)
	}

	return c, bps, nil
}

func parseConf(cfg []byte, cfgSection string) (conf, error) {
	// TODO: validate cfg for mandatory configuration. For example bpmon will
	// panic if influx.addr is not set.
	allSections := map[string]conf{}

	err := yaml.Unmarshal(cfg, &allSections)
	if err != nil {
		return conf{}, fmt.Errorf("Error while parsing configuration: %s", err.Error())
	}

	conf, ok := allSections[cfgSection]
	if !ok {
		return conf, fmt.Errorf("No section '%s' found in configuration", cfgSection)
	}

	//conf.Dashboard = config.GetDashboardConf(conf.Dashboard)
	//conf.Annotate = config.GetAnnotateConf(conf.Annotate)

	return conf, nil
}
