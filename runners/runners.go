package runners

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

type Runners map[string]Runner

func New(path string) (Runners, error) {
	r := Defaults()
	if path == "" {
		return r, nil
	}

	rdirs, err := ioutil.ReadDir(path)
	if err != nil {
		return r, fmt.Errorf("Error while reading runners from '%s': %s", path, err.Error())
	}

	for _, rdir := range rdirs {
		if !rdir.IsDir() {
			continue
		}

		metadatapath := fmt.Sprintf("%s/%s/cmd.yaml", path, rdir.Name())
		metadata, err := ioutil.ReadFile(metadatapath)
		if err != nil {
			return r, fmt.Errorf("Error while reading runner metadata (%s) for runner %s: %s", metadatapath, rdir.Name(), err.Error())
		}
		runner := Runner{}
		err = yaml.Unmarshal(metadata, &runner)
		if err != nil {
			return r, fmt.Errorf("Error while parsing runner metadata (%s) for runner %s: %s", metadatapath, rdir.Name(), err.Error())
		}

		templpath := fmt.Sprintf("%s/%s/cmd.template", path, rdir.Name())
		templfile, err := ioutil.ReadFile(templpath)
		if err != nil {
			return r, fmt.Errorf("Error while reading runner template (%s) for runner %s: %s", templpath, rdir.Name(), err.Error())
		}
		runner.Template, err = template.New(rdir.Name()).Parse(string(templfile))
		if err != nil {
			return r, fmt.Errorf("Error while parsing runner template (%s) for runner %s: %s", templpath, rdir.Name(), err.Error())
		}

		r[rdir.Name()] = runner
	}

	return r, nil
}

type Runner struct {
	Template    *template.Template `yaml:"template"`
	Description string             `yaml:"description"`
	Parameters  map[string]string  `yaml:"parameters"`
	ForEach     bool               `yaml:"for_each"`
}

func (r Runner) Exec(data interface{}) error {
	var command bytes.Buffer
	err := r.Template.Execute(&command, data)
	fmt.Print(command.String())
	return err
}
