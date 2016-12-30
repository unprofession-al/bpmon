package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	bpPath    string
	bpPattern string
	confPath  string
	i         icinga
)

type conf struct {
	Icinga icingaConf
}

type bps []bp

func init() {
	flag.StringVar(&bpPath, "bp", "conf/bp.d", "Path to business process definition files")
	flag.StringVar(&bpPattern, "pattern", "*.yaml", "File name pattern of business process definition files")
	flag.StringVar(&confPath, "conf", "conf/conf.yaml", "Path to config file")
}

func main() {
	flag.Parse()

	c, err := readConf()
	if err != nil {
		log.Fatal(err)
	}

	b, err := readBPs()
	if err != nil {
		log.Fatal(err)
	}

	i = newIcinga(c.Icinga)
	for _, bp := range b {
		rs := bp.Status()
		fmt.Println(rs.PrettyPrint(0))
	}
}

func readConf() (conf, error) {
	conf := conf{}
	file, err := ioutil.ReadFile(confPath)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error while reading %s: %s", confPath, err.Error()))
		return conf, err
	}

	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error while parsing %s: %s", confPath, err.Error()))
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
		bp := bp{}
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
