package main

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Configuration struct holds the global configuration for all parts
// of the App
type Configuration struct {
	Listener Listener `yaml:"listener"`
}

// Listener holds the configuration details for the web server
type Listener struct {
	Port    string `yaml:"port"`
	Address string `yaml:"address"`
}

// Configure reads a file from patAh `cfgFile`, unmarshalls its content
// to a Configuration and returns it,
func Configure(cfgFile string) (Configuration, error) {
	conf := Configuration{}

	cfgData, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return conf, fmt.Errorf("Error while reading configuration file '%s': %s", cfgFile, err.Error())
	}

	err = yaml.Unmarshal(cfgData, &conf)
	if err != nil {
		return conf, fmt.Errorf("Error while parsing configuration: %s", err.Error())
	}

	return conf, nil
}
