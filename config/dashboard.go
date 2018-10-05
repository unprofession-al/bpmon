package config

import "errors"

type DashboardConfig struct {
	Listener string `yaml:"listener"`
	Static   string `yaml:"static"`
}

func DashboardDefaults() DashboardConfig {
	return DashboardConfig{
		Listener: "127.0.0.1:8910",
	}
}

func (dc DashboardConfig) DashboardValidate() ([]string, error) {
	errs := []string{}
	if dc.Listener == "" {
		errs = append(errs, "Field 'listener' cannot be empty.")
	}
	if len(errs) > 0 {
		err := errors.New("Config of 'dashboard' has errors")
		return errs, err
	}
	return errs, nil
}

func (dc *DashboardConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Config Struct needs to be aliased in order to avoid an infinite recursiv
	// loop while unmarshalling
	type cd DashboardConfig
	out := cd(DashboardDefaults())
	err := unmarshal(&out)
	*dc = DashboardConfig(out)
	return err
}
