package runners

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"text/template"
	"time"

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
		return r, fmt.Errorf("error while reading runners from '%s': %s", path, err.Error())
	}

	for _, rdir := range rdirs {
		if !rdir.IsDir() {
			continue
		}

		metadatapath := fmt.Sprintf("%s/%s/cmd.yaml", path, rdir.Name())
		metadata, err := ioutil.ReadFile(metadatapath)
		if err != nil {
			return r, fmt.Errorf("error while reading runner metadata (%s) for runner %s: %s", metadatapath, rdir.Name(), err.Error())
		}
		runner := Runner{}
		err = yaml.Unmarshal(metadata, &runner)
		if err != nil {
			return r, fmt.Errorf("error while parsing runner metadata (%s) for runner %s: %s", metadatapath, rdir.Name(), err.Error())
		}

		templpath := fmt.Sprintf("%s/%s/cmd.template", path, rdir.Name())
		templfile, err := ioutil.ReadFile(templpath)
		if err != nil {
			return r, fmt.Errorf("error while reading runner template (%s) for runner %s: %s", templpath, rdir.Name(), err.Error())
		}
		runner.Template, err = template.New(rdir.Name()).Funcs(getFuncs()).Parse(string(templfile))
		if err != nil {
			return r, fmt.Errorf("error while parsing runner template (%s) for runner %s: %s", templpath, rdir.Name(), err.Error())
		}

		r[rdir.Name()] = runner
	}

	return r, nil
}

// AdHoc appends a new runner from a template string and returns its (generated) name/key
func (r *Runners) AdHoc(t string) (name string, err error) {
	name = fmt.Sprintf("adhoc-%s", strconv.FormatInt(time.Now().UTC().UnixNano(), 10))
	runner := Runner{}
	runner.Template, err = template.New(name).Funcs(getFuncs()).Parse(t)
	if err != nil {
		return name, fmt.Errorf("error while parsing runner template for: %s", err.Error())
	}
	(*r)[name] = runner
	return name, nil
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
