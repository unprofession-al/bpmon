package dashboard

import "errors"

type Config struct {
	Listener string `yaml:"listener"`
	Static   string `yaml:"static"`
}

func Defaults() Config {
	return Config{
		Listener: "127.0.0.1:8910",
	}
}

func (dc Config) Validate() ([]string, error) {
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
