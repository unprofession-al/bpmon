package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/periphery/annotate"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var (
	annotateAuthHeaderName   string
	annotateAuthHeaderValues string
)

var annotateCmd = &cobra.Command{
	Use:   "annotate",
	Short: "Run Web UI to annotate events",
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

		headerValues := strings.Split(annotateAuthHeaderValues, ",")
		for i, val := range headerValues {
			headerValues[i] = strings.TrimSpace(val)
		}

		router, err := annotate.Setup(s.Annotate, bp, pp, annotateAuthHeaderName, headerValues)
		if err != nil {
			msg := fmt.Sprintf("Could not build router for server: %s", err.Error())
			log.Fatal(msg)
		}

		fmt.Printf("Serving Annotaton UI at http://%s\nPress CTRL-c to stop...\n", s.Annotate.Listener)
		log.Fatal(http.ListenAndServe(s.Annotate.Listener, router))
	},
}

func init() {
	betaCmd.AddCommand(annotateCmd)
	annotateCmd.PersistentFlags().StringVarP(&annotateAuthHeaderName, "auth-header-name", "", "", "HTTP header name to check")
	annotateCmd.PersistentFlags().StringVarP(&annotateAuthHeaderValues, "auth-header-values", "", "", "HTTP header values to be allowed, comma separated list")
}
