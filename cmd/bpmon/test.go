package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/icinga"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run all business process checks and print to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		c, b, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			msg := fmt.Sprintf("Could not read section %s form file %s, error was %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		i, err := icinga.NewIcinga(c.Icinga, c.Rules)
		if err != nil {
			log.Fatal(err)
		}

		r := i.DefaultRules()
		r.Merge(c.Rules)
		if err != nil {
			log.Fatal(err)
		}

		infl, _ := bpmon.NewInflux(c.Influx)
		for _, bp := range b {
			rs := bp.Status(i, infl, r)
			rs.AddPreviousStatus(infl)
			fmt.Println(rs.PrettyPrint(0, printTimestamps, printValues))
		}
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
