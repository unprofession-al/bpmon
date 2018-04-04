package bpmon

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/configs"
	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/store"

	"gopkg.in/yaml.v2"
)

type conf struct {
	GlobalRecipient string                `yaml:"global_recipient"`
	Checker         checker.Conf          `yaml:"checker"`
	Store           store.Conf            `yaml:"store"`
	Availabilities  AvailabilitiesConf    `yaml:"availabilities"`
	Trigger         Trigger               `yaml:"trigger"`
	Health          Health                `yaml:"health"`
	Rules           rules.Rules           `yaml:"rules"`
	Dashboard       configs.DashboardConf `yaml:"dashboard"`
	Annotate        configs.AnnotateConf  `yaml:"annotate"`
}

type Trigger struct {
	Template string `yaml:"template"`
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

	conf.Dashboard = configs.GetDashboardConf(conf.Dashboard)
	conf.Annotate = configs.GetAnnotateConf(conf.Annotate)

	return conf, nil
}

func parseBP(bpconf []byte, a Availabilities, gr string) (BP, error) {
	bp := BP{}
	err := yaml.Unmarshal(bpconf, &bp)
	if err != nil {
		return bp, fmt.Errorf("Error while parsing: %s", err.Error())
	}

	if bp.AvailabilityName == "" {
		return bp, fmt.Errorf("There is no availability defined in business process config")
	}

	availability, ok := a[bp.AvailabilityName]
	if !ok {
		return bp, fmt.Errorf("The availability referenced '%s' does not exist", bp.AvailabilityName)
	}
	bp.Availability = availability

	if gr != "" {
		bp.Recipients = append(bp.Recipients, gr)
	}

	return bp, nil
}
