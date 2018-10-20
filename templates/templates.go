package templates

type Config map[string]Template

type Template struct {
	Template    string            `yaml:"template"`
	Description string            `yaml:"description"`
	Prompts     map[string]string `yaml:"prompts"`
}

/*
func (t Temlpates) Validate() ([]string, error) {
}
*/
