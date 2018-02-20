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

var (
	dashboardPepper           string
	dashboardRecipientsHeader string
	dashboardStatic           string
)

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

		if dashboardPepper != "" && dashboardRecipientsHeader != "" {
			log.Fatal("ERROR: pepper and recipients-header are, only one is allowed.")
		}
		if dashboardPepper == "" && dashboardRecipientsHeader == "" {
			fmt.Println("WARNING: No pepper or recipients-header is provided, all information are accessable without auth...")
		}

		recipientsHeaderName := dashboardRecipientsHeader
		authHeader := false
		if dashboardRecipientsHeader != "" {
			authHeader = true
			fmt.Printf("Recipients-header is provided, using %s to read recipients...\n", dashboardRecipientsHeader)
		}

		var recipientHashes map[string]string
		authPepper := false
		if dashboardPepper != "" {
			authPepper = true
			fmt.Println("Pepper is provided, generating auth hashes...")
			recipientHashes = bps.GenerateRecipientHashes(dashboardPepper)
			fmt.Printf("%15s: %s\n", "Recipient", "Hash")
			for k, v := range recipientHashes {
				fmt.Printf("%15s: %s\n", v, k)
			}
		}

		if dashboardStatic != "" {
			c.Dashboard.Static = dashboardStatic
		}

		router, err := dashboard.Setup(c.Dashboard, bps, pp, authPepper, recipientHashes, authHeader, recipientsHeaderName)
		if err != nil {
			msg := fmt.Sprintf("Could not build router for server: %s", err.Error())
			log.Fatal(msg)
		}

		fmt.Printf("Serving Dashboard at http://%s\nPress CTRL-c to stop...\n", c.Dashboard.Listener)
		log.Fatal(http.ListenAndServe(c.Dashboard.Listener, router))
	},
}

func init() {
	betaCmd.AddCommand(dashboardCmd)
	dashboardCmd.PersistentFlags().StringVarP(&dashboardPepper, "pepper", "", "", "Pepper used to generate auth token")
	dashboardCmd.PersistentFlags().StringVarP(&dashboardRecipientsHeader, "recipients-header", "", "", "HTTP header name to read recipients from")
	dashboardCmd.PersistentFlags().StringVarP(&dashboardStatic, "static", "", "", "Path to custom html frontend")
}
