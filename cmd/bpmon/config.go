package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the configurantion used to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := bpmon.ReadConf(cfgFile, cfgSection)
		if err != nil {
			log.Fatal(err)
		}

		var out []byte
		out, _ = yaml.Marshal(c)
		fmt.Println(string(out))
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
}
