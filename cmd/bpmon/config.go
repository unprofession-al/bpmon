package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon/checker"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the configurantion used to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.Load(cfgFile)
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

		s, err := c.Section(cfgSection)
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		i, err := checker.New(s.Checker)
		if err != nil {
			log.Fatal(err)
		}

		r := i.DefaultRules()
		err = r.Merge(s.Rules)
		if err != nil {
			log.Fatal(err)
		}

		s.Rules = r

		var out []byte
		out, _ = yaml.Marshal(s)
		fmt.Println(string(out))
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
}
