package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/icinga"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Insert data into influx db",
	Run: func(cmd *cobra.Command, args []string) {
		c, b, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			log.Fatal(err)
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
			log.Println("Processing " + bp.Name)
			rs := bp.Status(i, r)
			err = infl.Write(rs)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(writeCmd)
}
