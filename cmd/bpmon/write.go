package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/checker"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/persistence"
	_ "github.com/unprofession-al/bpmon/persistence/influx"
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

		i, err := checker.New(c.Checker)
		if err != nil {
			log.Fatal(err)
		}

		r := i.DefaultRules()
		r.Merge(c.Rules)
		if err != nil {
			log.Fatal(err)
		}

		p, _ := persistence.New(c.Persistence)
		for _, bp := range b {
			if verbose {
				log.Println("Processing " + bp.Name)
			}
			rs := bp.Status(i, p, r)
			if c.Persistence.GetLastStatus {
				rs.AddPreviousStatus(p, c.Persistence.SaveOK)
			}
			points := rs.GetPoints(c.Persistence.SaveOK)
			err = p.Write(points)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(writeCmd)
}
