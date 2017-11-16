package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	cfgSection string
	bpPath     string
	bpPattern  string
	verbose    bool
)

var RootCmd = &cobra.Command{
	Use:   "bpmon",
	Short: "Montior business processes composed of Icinga checks",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "cfg", "c", "/etc/bpmon/cfg.yaml", "path to the configuration file")
	RootCmd.PersistentFlags().StringVarP(&cfgSection, "section", "s", "default", "name of the section to be read")
	RootCmd.PersistentFlags().StringVarP(&bpPath, "bp", "b", "/etc/bpmon/bp.d", "path to business process configuration files")
	RootCmd.PersistentFlags().StringVarP(&bpPattern, "pattern", "p", "*.yaml", "pattern of business process configuration files to process")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", true, "print log output")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
