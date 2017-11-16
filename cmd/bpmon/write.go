package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/icinga"
)

var debug bool

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Insert data into InfluxDB",
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
			if verbose {
				log.Println("Processing " + bp.Name)
			}
			rs := bp.Status(i, infl, r)
			if c.Influx.GetLastStatus {
				rs.AddPreviousStatus(infl, c.Influx.SaveOK)
			}
			err = infl.Write(rs, debug)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(writeCmd)
	writeCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print InfluxDB line protocol instead of write to database")
}
