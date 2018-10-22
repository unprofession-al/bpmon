package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	"gopkg.in/yaml.v2"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Various tools to verify your BPMON config",
}

var configPrintCmd = &cobra.Command{
	Use:   "print",
	Short: "Print the given configurantion section as interpreted by BPMON to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := fmt.Sprintf("%s/%s", cfgBase, cfgFile)
		c, _, err := config.NewFromFile(cfg, injectDefaults)
		if err != nil {
			fmt.Println(err)
		}

		errs, err := c.Validate()
		if err != nil {
			for _, msg := range errs {
				fmt.Println(msg)
			}
			log.Fatal(err)
		}

		s, _, r, _, _, err := fromSection(c, cfgSection)
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		s.Rules = r

		var out []byte
		out, _ = yaml.Marshal(s)
		fmt.Println(string(out))
	},
}

var (
	configInitComments bool
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Print a configurantion with default values and comments to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		out := config.ExampleYAML(configInitComments)
		fmt.Println(string(out))
	},
}

var configRawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Print the configuration with its injected defaults to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := fmt.Sprintf("%s/%s", cfgBase, cfgFile)
		_, raw, err := config.NewFromFile(cfg, injectDefaults)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(raw))
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configPrintCmd)
	configCmd.AddCommand(configInitCmd)
	configInitCmd.PersistentFlags().BoolVarP(&configInitComments, "comments", "", true, "print documentation comments")
	configCmd.AddCommand(configRawCmd)
}
