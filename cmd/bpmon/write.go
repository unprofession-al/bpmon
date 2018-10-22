package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Insert data into InfluxDB",
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

		s, i, r, b, p, err := fromSection(c, cfgSection)
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}
		for _, bp := range b {
			if verbose {
				log.Println("Processing " + bp.Name)
			}
			rs := bp.Status(i, p, r)
			if s.Store.GetLastStatus {
				rs.AddPreviousStatus(p, s.Store.SaveOK)
			}
			err = p.Write(&rs)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(writeCmd)
}
