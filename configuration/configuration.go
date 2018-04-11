package configuration

import (
	"fmt"
	"io/ioutil"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

type Config map[string]Section

type Validater interface {
	Validate() (error, []string)
}

type Section struct {
	Health HealthConfig `yaml:"health"`
}

func (s Section) Validate() error {
	v := reflect.ValueOf(s)

	for i := 0; i < v.NumField(); i++ {
		//fragment := v.Field(i)

	}

	return nil
}

var c = make(Config)

func Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error while reading configuration file '%s': %s", path, err.Error())
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("Error while unmarshalling configuration from yaml: %s", err.Error())
	}
	return nil
}

func GetAll() Config {
	return c
}

func GetSection(name string) Section {
	if section, ok := c[name]; ok {
		return section
	}
	return Section{}
}
