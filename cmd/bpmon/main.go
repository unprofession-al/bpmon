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
