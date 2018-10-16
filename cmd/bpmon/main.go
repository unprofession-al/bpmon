package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/rules"
	"github.com/unprofession-al/bpmon/store"
)

var (
	cfgFile        string
	cfgSection     string
	bpPath         string
	bpPattern      string
	verbose        bool
	injectDefaults bool
)

var RootCmd = &cobra.Command{
	Use:   "bpmon",
	Short: "Montior business processes composed of Icinga checks",
}

var betaCmd = &cobra.Command{
	Use:   "beta",
	Short: "Access beta features of BPMON",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "cfg", "c", "/etc/bpmon/cfg.yaml", "path to the configuration file")
	RootCmd.PersistentFlags().StringVarP(&cfgSection, "section", "s", "default", "name of the section to be read")
	RootCmd.PersistentFlags().StringVarP(&bpPath, "bp", "b", "/etc/bpmon/bp.d", "path to business process configuration files")
	RootCmd.PersistentFlags().StringVarP(&bpPattern, "pattern", "p", "*.yaml", "pattern of business process configuration files to process")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", true, "print log output")
	RootCmd.PersistentFlags().BoolVarP(&injectDefaults, "defaults", "d", true, "inject defaults in main config file")
	RootCmd.AddCommand(betaCmd)
}

func fromSection(cnf config.Config, sectionName string) (s config.ConfigSection, c checker.Checker, r rules.Rules, b bpmon.BusinessProcesses, p store.Accessor, err error) {
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

	b, err = bpmon.LoadBP(bpPath, bpPattern, a, s.GlobalRecipient)
	if err != nil {
		return
	}

	p, err = store.New(s.Store)
	return
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
