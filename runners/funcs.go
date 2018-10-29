package runners

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/pretty"
	yaml "gopkg.in/yaml.v2"
)

func getFuncs() template.FuncMap {
	return template.FuncMap{
		"pretty": func(i interface{}) string {
			return pretty.Sprint(i)
		},
		"json": func(i interface{}) string {
			json, _ := json.MarshalIndent(i, "", "\t")
			return string(json)
		},
		"yaml": func(i interface{}) string {
			yaml, _ := yaml.Marshal(i)
			return string(yaml)
		},
		"spew": func(i interface{}) string {
			return spew.Sprint(i)
		},
		"describe": func(i interface{}) string {
			return describeStruct(i, 0)
		},
	}
}

func describeStruct(t interface{}, depth int) string {
	prefix := strings.Repeat("  ", depth)
	var out string
	s := reflect.Indirect(reflect.ValueOf(t))
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		out = fmt.Sprintf("%s%s%s %s\n", out, prefix, typeOfT.Field(i).Name, typeOfT.Field(i).Type)
		switch f.Type().Kind() {
		case reflect.Struct, reflect.Ptr:
			out = fmt.Sprintf("%s%s{\n", out, prefix)
			out = fmt.Sprintf("%s%s", out, describeStruct(f.Interface(), depth+1))
			out = fmt.Sprintf("%s%s}\n", out, prefix)
		}
	}
	return out
}
