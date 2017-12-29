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

var dashboardPepper string

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Run Dashboard Web UI",
	Run: func(cmd *cobra.Command, args []string) {
		c, bps, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			msg := fmt.Sprintf("Could not read section %s form file %s, error was %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		pp, _ := store.New(c.Store)

		var recipientHashes map[string]string
		auth := false
		if dashboardPepper != "" {
			auth = true
			fmt.Println("Pepper is provided, generating auth hashes...")
			recipientHashes = bps.GenerateHashes(dashboardPepper)
			fmt.Printf("%15s: %s\n", "Recipient", "Hash")
			for k, v := range recipientHashes {
				fmt.Printf("%15s: %s\n", v, k)
			}
		} else {
			fmt.Println("WARNING: No pepper is provided, all information are accessable without auth...")
		}

		router, err := dashboard.Setup(c.Dashboard, bps, pp, auth, recipientHashes)
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
	dashboardCmd.PersistentFlags().StringVarP(&dashboardPepper, "pepper", "", "", "Pepper used to generate auth token")
}
