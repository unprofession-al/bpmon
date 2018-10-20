package templates

type Config map[string]Template

type Template struct {
	Template    string            `yaml:"template"`
	Description string            `yaml:"description"`
	Parameters  map[string]string `yaml:"parameters"`
}

/*
func (t Temlpates) Validate() ([]string, error) {
}
*/
