package trigger

import "errors"

type Config struct {
	Template string `yaml:"template"`
}

func Defaults() Config {
	return Config{
		Template: "{{.}}",
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
