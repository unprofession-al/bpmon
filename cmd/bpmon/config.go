package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/checker"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the configurantion used to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		c, _, err := bpmon.Configure(cfgFile, cfgSection, "", "")
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		i, err := checker.New(c.Checker)
		if err != nil {
			log.Fatal(err)
		}

		r := i.DefaultRules()
		err = r.Merge(c.Rules)
		if err != nil {
			log.Fatal(err)
		}

		c.Rules = r

		var out []byte
		out, _ = yaml.Marshal(c)
		fmt.Println(string(out))
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
}
