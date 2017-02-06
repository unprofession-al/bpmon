package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type conf struct {
	Icinga         IcingaConf         `yaml:"icinga"`
	Influx         InfluxConf         `yaml:"influx"`
	Availabilities AvailabilitiesConf `yaml:"availabilities"`
}

type bps []BP

func configure() (conf, bps, error) {
	c, err := readConf()
	if err != nil {
		return c, nil, err
	}

	a, err := c.Availabilities.Parse()
	if err != nil {
		return c, nil, err
	}

	b, err := readBPs(a)
	if err != nil {
		return c, nil, err
	}

	return c, b, nil
}

func readConf() (conf, error) {
	// TODO: validate cfg for mandatory configuration. For example bpmon will
	// panic if influx.addr is not set.
	conf := conf{}
	file, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error while reading %s: %s", cfgFile, err.Error()))
		return conf, err
	}

	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error while parsing %s: %s", cfgFile, err.Error()))
		return conf, err
	}

	return conf, nil
}

func readBPs(a Availabilities) (bps, error) {
	bps := bps{}
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
