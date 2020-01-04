package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon/internal/config"
	"github.com/unprofession-al/bpmon/internal/dashboard"
	_ "github.com/unprofession-al/bpmon/internal/store/influx"
)

var (
	dashboardPepper string
	dashboardHeader string
	dashboardStatic string
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Run Dashboard Web UI",
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

		s, _, _, bp, store, err := fromSection(c, cfgSection)
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		if dashboardStatic != "" {
			s.Dashboard.Static = dashboardStatic
		}

		d, msg, err := dashboard.New(s.Dashboard, bp, store, dashboardPepper, dashboardHeader)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(msg)

		d.Run()
	},
}

func init() {
	RootCmd.AddCommand(dashboardCmd)
	dashboardCmd.PersistentFlags().StringVarP(&dashboardPepper, "pepper", "", "", "Pepper used to generate auth token")
	dashboardCmd.PersistentFlags().StringVarP(&dashboardHeader, "header", "", "", "HTTP header name to read recipients from")
	dashboardCmd.PersistentFlags().StringVarP(&dashboardStatic, "static", "", "", "Path to custom html frontend")
}
