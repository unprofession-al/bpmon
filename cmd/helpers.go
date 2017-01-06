package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type conf struct {
	Icinga IcingaConf
	Influx InfluxConf
}

type bps []BusinessProcess

func configure() (conf, bps, error) {
	c, err := readConf()
	if err != nil {
		return c, nil, err
	}

	b, err := readBPs()
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

func readBPs() (bps, error) {
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
		bp := BusinessProcess{}
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

		bps = append(bps, bp)
	}
	return bps, nil
}
