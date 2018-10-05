package config

import "errors"

type AnnotateConfig struct {
	Listener string `yaml:"listener"`
	Static   string `yaml:"static"`
}

func AnnotateDefaults() AnnotateConfig {
	return AnnotateConfig{
		Listener: "127.0.0.1:8765",
	}
}

func (ac AnnotateConfig) AnnotateValidate() ([]string, error) {
	errs := []string{}
	if ac.Listener == "" {
		errs = append(errs, "Field 'listener' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'annotate' has errors")
		return errs, err
	}
	return errs, nil
}

func (ac *AnnotateConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Config Struct needs to be aliased in order to avoid an infinite recursiv
	// loop while unmarshalling
	type cd AnnotateConfig
	out := cd(AnnotateDefaults())
	err := unmarshal(&out)
	*ac = AnnotateConfig(out)
	return err
}
