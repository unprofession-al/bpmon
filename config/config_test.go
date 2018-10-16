package config

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestDefaultedConfig(t *testing.T) {
	c, _, err := New(InputData, true)
	if err != nil {
		t.Errorf("Loading string failed with error: %s", err.Error())
	}

	for section, data := range ExpectedResultsDefaulted {
		t.Run(section, func(t *testing.T) {

			if diff := pretty.Compare(data, c[section]); diff != "" {
				t.Errorf("Comparing section '%s' failed (-got +want):\n%s", section, diff)
			}
		})
	}
}

var InputData = []byte(`---
defaults:
`)

var ExpectedResultsDefaulted = Config{
	"defaults": defaultConfigSection(),
}
