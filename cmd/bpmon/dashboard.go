package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/periphery/dashboard"
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
		c, err := config.Load(cfgFile)
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

		s, _, _, bp, pp, err := fromSection(c, cfgSection)
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		if dashboardPepper != "" && dashboardRecipientsHeader != "" {
			log.Fatal("ERROR: pepper and recipients-header are set, only one is allowed.")
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
			recipientHashes = bp.GenerateRecipientHashes(dashboardPepper)
			fmt.Printf("%15s: %s\n", "Recipient", "Hash")
			for k, v := range recipientHashes {
				fmt.Printf("%15s: %s\n", v, k)
			}
		}

		if dashboardStatic != "" {
			s.Dashboard.Static = dashboardStatic
		}

		router, err := dashboard.Setup(s.Dashboard, bp, pp, authPepper, recipientHashes, authHeader, recipientsHeaderName)
		if err != nil {
			msg := fmt.Sprintf("Could not build router for server: %s", err.Error())
			log.Fatal(msg)
		}

		fmt.Printf("Serving Dashboard at http://%s\nPress CTRL-c to stop...\n", s.Dashboard.Listener)
		log.Fatal(http.ListenAndServe(s.Dashboard.Listener, router))
	},
}

func init() {
	betaCmd.AddCommand(dashboardCmd)
	dashboardCmd.PersistentFlags().StringVarP(&dashboardPepper, "pepper", "", "", "Pepper used to generate auth token")
	dashboardCmd.PersistentFlags().StringVarP(&dashboardRecipientsHeader, "recipients-header", "", "", "HTTP header name to read recipients from")
	dashboardCmd.PersistentFlags().StringVarP(&dashboardStatic, "static", "", "", "Path to custom html frontend")
}
