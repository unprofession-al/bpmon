package bpmon

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type conf struct {
	Icinga         IcingaConf         `yaml:"icinga"`
	Influx         InfluxConf         `yaml:"influx"`
	Availabilities AvailabilitiesConf `yaml:"availabilities"`
	Trigger        Trigger            `yaml:"trigger"`
	Rules          []Rule             `yaml:"rules"`
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
