package cmd

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
)

var RootCmd = &cobra.Command{
	Use:   "bpmon",
	Short: "Montior business processes composed of Icinga checks",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "cfg", "c", "/etc/bpmon/cfg.yaml", "config file (default is \"/etc/bpmon/cfg.yaml\")")
	RootCmd.PersistentFlags().StringVarP(&cfgSection, "section", "s", "default", "Which section to be read")
	RootCmd.PersistentFlags().StringVarP(&bpPath, "bp", "b", "/etc/bpmon/bp.d", "path to business process config files")
	RootCmd.PersistentFlags().StringVarP(&bpPattern, "pattern", "p", "*.yaml", "pattern of business process config files to process")
}
