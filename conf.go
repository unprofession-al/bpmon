package bpmon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/configs"
	"github.com/unprofession-al/bpmon/persistence"
	"github.com/unprofession-al/bpmon/rules"

	"gopkg.in/yaml.v2"
)

type conf struct {
	Checker        checker.Conf          `yaml:"checker"`
	Persistence    persistence.Conf      `yaml:"persistence"`
	Availabilities AvailabilitiesConf    `yaml:"availabilities"`
	Trigger        Trigger               `yaml:"trigger"`
	Rules          rules.Rules           `yaml:"rules"`
	Dashboard      configs.DashboardConf `yaml:"dashboard"`
}

type Trigger struct {
	Template string `yaml:"template"`
}

func Configure(cfgFile, cfgSection, bpPath, bpPattern string) (conf, BusinessProcesses, error) {
	cfgData, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return conf{}, nil, errors.New(fmt.Sprintf("Error while reading configuration file '%s': %s", cfgFile, err.Error()))
	}

	c, err := parseConf(cfgData, cfgSection)
	if err != nil {
		return c, nil, errors.New(fmt.Sprintf("Error while parsing section '%s' configuration file '%s': %s", cfgSection, cfgFile, err.Error()))
	}

	a, err := c.Availabilities.Parse()
	if err != nil {
		return c, nil, errors.New(fmt.Sprintf("Error while parsing availabilities from configuration file '%s': %s", cfgFile, err.Error()))
	}

	var bps BusinessProcesses
	if bpPath == "" {
		return c, bps, nil
	}
	files, err := ioutil.ReadDir(bpPath)
	if err != nil {
		return c, bps, errors.New(fmt.Sprintf("Error while reading business configuration files from '%s': %s", bpPath, err.Error()))
	}
	for _, f := range files {
		match, err := filepath.Match(bpPattern, f.Name())
		if err != nil {
			return c, bps, errors.New(fmt.Sprintf("Error while matching file pattern '%s' in '%s': %s", bpPattern, bpPath, err.Error()))
		}
		if !match {
			continue
		}
		file, err := ioutil.ReadFile(bpPath + "/" + f.Name())
		if err != nil {
			return c, bps, errors.New(fmt.Sprintf("Error while reading business process %s/%s: %s", bpPath, f.Name(), err.Error()))
		}
		bp, err := parseBP(file, a)
		if err != nil {
			return c, bps, errors.New(fmt.Sprintf("Error while parsing business process%s/%s: %s", bpPath, f.Name(), err.Error()))
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
		return conf{}, errors.New(fmt.Sprintf("Error while parsing configuration: %s", err.Error()))
	}

	conf, ok := allSections[cfgSection]
	if !ok {
		return conf, errors.New(fmt.Sprintf("No section '%s' found in configuration", cfgSection))
	}

	conf.Dashboard = configs.GetDashboardConf(conf.Dashboard)

	return conf, nil
}

func parseBP(bpconf []byte, a Availabilities) (BP, error) {
	bp := BP{}
	err := yaml.Unmarshal(bpconf, &bp)
	if err != nil {
		return bp, errors.New(fmt.Sprintf("Error while parsing: %s", err.Error()))
	}

	if bp.AvailabilityName == "" {
		return bp, errors.New(fmt.Sprintf("There is no availability defined in business process config"))
	}

	availability, ok := a[bp.AvailabilityName]
	if !ok {
		return bp, errors.New(fmt.Sprintf("The availability referenced '%s' does not exist", bp.AvailabilityName))
	}
	bp.Availability = availability

	return bp, nil
}
