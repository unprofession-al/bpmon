package dashboard

import "errors"

type Config struct {
	// listener tells the dashboard where to bind. This string
	// should match the pattern [ip]:[port].
	Listener string `yaml:"listener"`

	// static is the path to the directory that sould be served
	// at the root of the server. This should contain the UI of the
	// Dashboard
	Static string `yaml:"static"`

	// grant_write is a list of recipients which are allowed to access the annotate
	// endpoint via POST request.
	GrantWrite []string `yaml:"grant_write"`
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
