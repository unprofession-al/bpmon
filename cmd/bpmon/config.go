package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
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

func init() {
	RootCmd.AddCommand(configCmd)
}
