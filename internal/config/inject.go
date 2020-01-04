package config

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func injectDefaults(data []byte, key string) ([]byte, error) {
	// render a default section an add an anchor
	defaultData := Config{key: defaultConfigSection()}
	defaultSection, _ := yaml.Marshal(defaultData)
	defaultSection = bytes.Replace(defaultSection, []byte(key+":"), []byte(key+": &"+key), 1)

	// get list of sections in input data
	c := Config{}
	err := yaml.Unmarshal(data, &c)
	if err != nil {
		return data, fmt.Errorf("error while reading sections from yaml: %s", err.Error())
	}

	// remove "---" at beginning when present
	data = bytes.TrimLeft(data, "---")

	// add reference to default section to each section
	lines := bytes.Split(data, []byte("\n"))
	var updatedLines [][]byte
	for _, line := range lines {
		updatedLines = append(updatedLines, line)
		for section := range c {
			if bytes.HasPrefix(line, []byte(section+":")) {
				updatedLines = append(updatedLines, []byte("  <<: *"+key))
			}
		}
	}
	updatedData := bytes.Join(updatedLines, []byte("\n"))

	// compose injected yaml
	out := []byte("---\n")
	out = append(out, defaultSection...)
	out = append(out, updatedData...)
	return out, nil
}

var indent = regexp.MustCompile(`^(?P<ws>\s*)(?P<key>\w+):.*`)

func injectComments(data []byte, section string) []byte {
	comments := configDoc(section)
	indentString := getSingleIndent(data)
	depth := 0
	var node [15]string

	lines := bytes.Split(data, []byte("\n"))

	var updatedLines [][]byte
	for _, line := range lines {
		if matches := indent.FindStringSubmatch(string(line)); len(matches) > 2 {
			// wee need to add comments
			currentIndent := matches[1]
			currentKey := matches[2]
			depth = (len(currentIndent) / len(indentString))
			node[depth] = currentKey
			fullKey := strings.Join(node[:depth+1], ".")
			if comment, ok := comments[fullKey]; ok {
				for _, cl := range strings.Split(comment, "\n") {
					if cl != "" {
						indentedComment := fmt.Sprintf("%s# %s", strings.Repeat(string(indentString), depth), cl)
						updatedLines = append(updatedLines, []byte(indentedComment))
					}
				}
			}
		}
		updatedLines = append(updatedLines, line)
	}
	updatedData := bytes.Join(updatedLines, []byte("\n"))

	// compose injected yaml
	out := []byte("")
	out = append(out, updatedData...)
	return out
}

var whitespaces = regexp.MustCompile(`^(?P<ws>\s+).*`)

func getSingleIndent(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		matches := whitespaces.FindStringSubmatch(string(line))
		if len(matches) > 1 {
			return []byte(matches[1])
		}
	}
	return []byte{}
}
