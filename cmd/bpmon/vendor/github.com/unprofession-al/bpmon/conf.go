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
	c, err := ReadConf(cfgFile, cfgSection)
	if err != nil {
		return c, nil, err
	}

	a, err := c.Availabilities.Parse()
	if err != nil {
		return c, nil, err
	}

	b, err := readBPs(bpPath, bpPattern, a)
	if err != nil {
		return c, nil, err
	}

	return c, b, nil
}

func ReadConf(cfgFile, cfgSection string) (conf, error) {
	// TODO: validate cfg for mandatory configuration. For example bpmon will
	// panic if influx.addr is not set.
	allSections := map[string]conf{}
	conf := conf{}
	file, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return conf, errors.New(fmt.Sprintf("Error while reading %s: %s", cfgFile, err.Error()))
	}

	err = yaml.Unmarshal(file, &allSections)
	if err != nil {
		return conf, errors.New(fmt.Sprintf("Error while parsing %s: %s", cfgFile, err.Error()))
	}

	conf, ok := allSections[cfgSection]
	if !ok {
		return conf, errors.New(fmt.Sprintf("No section '%s' found in file %s", cfgSection, cfgFile))
	}

	return conf, nil
}

func readBPs(bpPath, bpPattern string, a Availabilities) (BusinessProcesses, error) {
	bps := BusinessProcesses{}

	files, err := ioutil.ReadDir(bpPath)
	if err != nil {
		return bps, err
	}

	for _, f := range files {
		match, err := filepath.Match(bpPattern, f.Name())
		if err != nil {
			return bps, err
		}
		if !match {
			continue
		}
		bp := BP{}
		file, err := ioutil.ReadFile(bpPath + "/" + f.Name())
		if err != nil {
			err = errors.New(fmt.Sprintf("Error while reading %s/%s: %s", bpPath, f.Name(), err.Error()))
			return bps, err
		}

		err = yaml.Unmarshal(file, &bp)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error while parsing %s/%s: %s", bpPath, f.Name(), err.Error()))
			return bps, err
		}

		if bp.AvailabilityName == "" {
			err = errors.New(fmt.Sprintf("There is no availability defined in %s/%s", bpPath, f.Name()))
			return bps, err
		}

		availability, ok := a[bp.AvailabilityName]
		if !ok {
			err = errors.New(fmt.Sprintf("The availability '%s' referenced in '%s/%s' does not exist", bp.AvailabilityName, bpPath, f.Name()))
			return bps, err
		}
		bp.Availability = availability

		bps = append(bps, bp)
	}
	return bps, nil
}
