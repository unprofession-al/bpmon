package configuration

import "errors"

type HealthConfig struct {
	Template        string            `yaml:"template"`
	StoreRequired   bool              `yaml:"store_required"`
	CheckerRequired bool              `yaml:"checker_required"`
	Responsible     string            `yaml:"responsible"`
	Name            string            `yaml:"name"`
	ID              string            `yaml:"id"`
	Bla             map[string]string `yaml:"bla"`
}

func (hc HealthConfig) Validate() (error, []string) {
	errs := []string{}
	if hc.Template == "" {
		errs = append(errs, "Field 'template' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'health' has errors")
		return err, errs
	}
	return nil, errs
}

func (hc *HealthConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type HealthConfigDefaulted HealthConfig

	var defaults = HealthConfigDefaulted{
		Template:        "{{.}}",
		StoreRequired:   false,
		CheckerRequired: true,
		Responsible:     "",
		Name:            "",
		ID:              "bla",
	}

	out := defaults
	err := unmarshal(&out)
	*hc = HealthConfig(out)
	return err
}
