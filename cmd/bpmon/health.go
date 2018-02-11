package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/checker"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health of store and checker",
	Run: func(cmd *cobra.Command, args []string) {
		c, _, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			msg := fmt.Sprintf("Could not read section %s form file %s, error was %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		ch, err := checker.New(c.Checker)
		if err != nil {
			log.Fatal(err)
		}

		st, err := store.New(c.Store)
		if err != nil {
			log.Fatal(err)
		}

		_, checkerErr := ch.Health()
		_, storeErr := st.Health()

		if checkerErr != nil && storeErr != nil {
			log.Fatal("dependencies failed")
		}

	},
}

func init() {
	RootCmd.AddCommand(healthCmd)
}
