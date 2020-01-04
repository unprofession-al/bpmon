package main

import (
	"fmt"
	"os"

	"github.com/unprofession-al/bpmon/internal/bpmon"
	"github.com/unprofession-al/bpmon/internal/checker"
	"github.com/unprofession-al/bpmon/internal/config"
	"github.com/unprofession-al/bpmon/internal/rules"
	"github.com/unprofession-al/bpmon/internal/store"
)

var (
	// These variables are passed during `go build` via ldflags, for example:
	//   go build -ldflags "-X main.commit=$(git rev-list -1 HEAD)"
	// goreleaser (https://goreleaser.com/) does this by default.
	version string
	commit  string
	date    string
)

func main() {
	if err := NewApp().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func fromSection(cnf config.Config, sectionName, cfgBase, bpPattern string) (s config.ConfigSection, c checker.Checker, r rules.Rules, b bpmon.BusinessProcesses, p store.Accessor, err error) {
	s, err = cnf.Section(sectionName)
	if err != nil {
		return
	}

	c, err = checker.New(s.Checker)
	if err != nil {
		return
	}

	r = c.DefaultRules()
	err = r.Merge(s.Rules)
	if err != nil {
		return
	}

	a, err := s.Availabilities.Parse()
	if err != nil {
		return
	}

	bpPath := fmt.Sprintf("%s/%s", cfgBase, s.Env.BP)
	b, err = bpmon.LoadBP(bpPath, bpPattern, a, s.GlobalRecipients)
	if err != nil {
		return
	}

	p, err = store.New(s.Store)
	return
}

// verisonInfo returns a string containing information usually passed via
// ldflags during build time.
func versionInfo() string {
	if version == "" {
		version = "dirty"
	}
	if commit == "" {
		commit = "dirty"
	}
	if date == "" {
		date = "unknown"
	}
	return fmt.Sprintf("Version:    %s\nCommit:     %s\nBuild Date: %s\n", version, commit, date)
}
