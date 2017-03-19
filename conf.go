package bpmon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/unprofession-al/bpmon/icinga"
	"github.com/unprofession-al/bpmon/rules"

	"gopkg.in/yaml.v2"
)

type conf struct {
	Icinga         icinga.IcingaConf  `yaml:"icinga"`
	Influx         InfluxConf         `yaml:"influx"`
	Availabilities AvailabilitiesConf `yaml:"availabilities"`
	Trigger        Trigger            `yaml:"trigger"`
	Rules          rules.Rules        `yaml:"rules"`
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
		return c, nil, err
	}

	a, err := c.Availabilities.Parse()
	if err != nil {
		return c, nil, err
	}

	var bps BusinessProcesses
	if bpPath == "" {
		return c, bps, nil
	}
	files, err := ioutil.ReadDir(bpPath)
	if err != nil {
		return c, bps, err
	}
	for _, f := range files {
		match, err := filepath.Match(bpPattern, f.Name())
		if err != nil {
			return c, bps, err
		}
		if !match {
			continue
		}
		file, err := ioutil.ReadFile(bpPath + "/" + f.Name())
		if err != nil {
			err = errors.New(fmt.Sprintf("Error while reading %s/%s: %s", bpPath, f.Name(), err.Error()))
			return c, bps, err
		}
		bp, err := parseBP(file, a)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error while parsing %s/%s: %s", bpPath, f.Name(), err.Error()))
			return c, bps, err
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

	return conf, nil
}

func parseBP(bpconf []byte, a Availabilities) (BP, error) {
	bp := BP{}
	err := yaml.Unmarshal(bpconf, &bp)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error while parsing: %s", err.Error()))
		return bp, err
	}

	if bp.AvailabilityName == "" {
		err = errors.New(fmt.Sprintf("There is no availability defined in business process config"))
		return bp, err
	}

	availability, ok := a[bp.AvailabilityName]
	if !ok {
		err = errors.New(fmt.Sprintf("The availability referenced '%s' does not exist", bp.AvailabilityName))
		return bp, err
	}
	bp.Availability = availability

	return bp, nil
}
