package trigger

import "text/template"

type Trigger struct {
	Template *template.Template `yaml:"template"`
}

func New(c Config) (Trigger, error) {
	templ, err := template.New("t1").Parse(c.Template)
	t := Trigger{
		Template: templ,
	}
	return t, err
}
