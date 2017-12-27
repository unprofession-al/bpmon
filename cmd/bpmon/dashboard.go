package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/periphery/dashboard"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Run all business process checks and print to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		c, bp, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			msg := fmt.Sprintf("Could not read section %s form file %s, error was %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		pp, _ := store.New(c.Store)

		router, err := dashboard.Setup(c.Dashboard, bp, pp)
		if err != nil {
			msg := fmt.Sprintf("Could not build router for server: %s", err.Error())
			log.Fatal(msg)
		}

		listen := fmt.Sprintf("%s:%d", c.Dashboard.Address, c.Dashboard.Port)
		fmt.Printf("Serving Dashboard at http://%s\nPress CTRL-c to stop...\n", listen)
		log.Fatal(http.ListenAndServe(listen, router))
	},
}

func init() {
	betaCmd.AddCommand(dashboardCmd)
}
