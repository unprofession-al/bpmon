package config

import "errors"

type EnvConfig struct {
	Runners string `yaml:"runner"`
	BP      string `yaml:"bp"`
}

func EnvDefaults() EnvConfig {
	return EnvConfig{
		Runners: "runners/",
		BP:      "bp.d/",
	}
}

func (env EnvConfig) Validate() ([]string, error) {
	errs := []string{}
	if env.Runners == "" {
		errs = append(errs, "Field 'runner' cannot be empty.")
	}
	if env.BP == "" {
		errs = append(errs, "Field 'bp' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'env' has errors")
		return errs, err
	}
	return errs, nil
}
