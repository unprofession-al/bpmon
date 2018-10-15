package health

import "errors"

type Config struct {
	Template        string `yaml:"template"`
	StoreRequired   bool   `yaml:"store_required"`
	CheckerRequired bool   `yaml:"checker_required"`
	Responsible     string `yaml:"responsible"`
	Name            string `yaml:"name"`
	ID              string `yaml:"id"`
}

func Defaults() Config {
	return Config{
		Template:        "{{.}}",
		StoreRequired:   false,
		CheckerRequired: true,
		Responsible:     "",
		Name:            "",
		ID:              "bla",
	}
}

func (c Config) Validate() ([]string, error) {
	errs := []string{}
	if c.Template == "" {
		errs = append(errs, "Field 'template' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'health' has errors")
		return errs, err
	}
	return errs, nil
}
