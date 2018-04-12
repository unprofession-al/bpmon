package configuration

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config map[string]Section

type Section struct {
	Health HealthConfig `yaml:"health"`
}

func Load(path string) (Config, error) {
	c := Config{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("Error while reading configuration file '%s': %s", path, err.Error())
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, fmt.Errorf("Error while unmarshalling configuration from yaml: %s", err.Error())
	}
	return c, nil
}

func (s Section) Validate() (error, []string) {
	return s.Health.Validate()
}

func (c Config) Section(name string) Section {
	if section, ok := c[name]; ok {
		return section
	}
	return Section{}
}
