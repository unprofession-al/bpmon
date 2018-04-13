package bpmon

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Conf map[string]ConfSection

type ConfSection struct {
	Health HealthConfig `yaml:"health"`
}

func Load(path string) (Conf, error) {
	c := ConfDefaults()

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

func ConfDefaults() Conf {
	c := make(Conf)
	c["default"] = ConfSection{
		Health: HealthConfigDefaults(),
	}
	return c
}

func (s ConfSection) Validate() ([]string, error) {
	return s.Health.Validate()
}

func (c Conf) Section(name string) ConfSection {
	if section, ok := c[name]; ok {
		return section
	}
	return ConfSection{}
}
