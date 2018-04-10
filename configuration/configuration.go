package configuration

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

const DefaultSectionName = "default"

var unmarshallers = make(map[string]func(interface{}) (Fragment, error))

type Section map[string]Fragment
type Config map[string]Section

func Register(k string, u func(interface{}) (Fragment, error)) error {
	if _, ok := unmarshallers[k]; ok {
		return fmt.Errorf("Key %s already exists", k)
	}
	unmarshallers[k] = u
	return nil
}

var c = Config{}

func Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error while reading configuration file '%s': %s", path, err.Error())
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("Error while parsing configuration: %s", err.Error())
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
	return nil
}

func (section *Section) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux map[string]interface{}
	if err := unmarshal(&aux); err != nil {
		return err
	}

	section = &Section{}
	for key, value := range aux {
		if unmarshaller, ok := unmarshallers[key]; ok {
			fragment, err := unmarshaller(value)
			if err != nil {
				return err
			}
			(*section)[key] = fragment
		} else {
			return fmt.Errorf("Unknown configuration fragment %s", key)
		}
	}
	fmt.Println(*section)
	return nil
}
