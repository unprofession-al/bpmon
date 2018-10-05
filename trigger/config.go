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

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Config Struct needs to be aliased in order to avoid an infinite recursiv
	// loop while unmarshalling
	type cd Config
	out := cd(Defaults())
	err := unmarshal(&out)
	*c = Config(out)
	return err
}
