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

const baseEnv = "BPMON_BASE"

var (
	cfgFile        string
	cfgBase        string
	cfgSection     string
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
	base := os.Getenv(baseEnv)
	if base == "" {
		base = "."
	}

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "name the configuration file")
	RootCmd.PersistentFlags().StringVarP(&cfgBase, "base", "b", base, fmt.Sprintf("path of the directory where the configuration is located (default can be set via $%s)", baseEnv))
	RootCmd.PersistentFlags().StringVarP(&cfgSection, "section", "s", "default", "name of the section to be read")
	RootCmd.PersistentFlags().StringVarP(&bpPattern, "pattern", "p", "*.yaml", "pattern of business process configuration files to process")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", true, "print log output")
	RootCmd.PersistentFlags().BoolVarP(&injectDefaults, "defaults", "d", true, "inject defaults in main config file")
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

	bpPath := fmt.Sprintf("%s/%s", cfgBase, s.Env.BP)
	b, err = bpmon.LoadBP(bpPath, bpPattern, a, s.GlobalRecipients)
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
