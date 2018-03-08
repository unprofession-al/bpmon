package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/periphery/annotate"
	"github.com/unprofession-al/bpmon/store"
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
		c, bp, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			msg := fmt.Sprintf("Could not read section %s form file %s, error was %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		pp, _ := store.New(c.Store)

		headerValues := strings.Split(annotateAuthHeaderValues, ",")
		for i, val := range headerValues {
			headerValues[i] = strings.TrimSpace(val)
		}

		router, err := annotate.Setup(c.Annotate, bp, pp, annotateAuthHeaderName, headerValues)
		if err != nil {
			msg := fmt.Sprintf("Could not build router for server: %s", err.Error())
			log.Fatal(msg)
		}

		fmt.Printf("Serving Annotaton UI at http://%s\nPress CTRL-c to stop...\n", c.Annotate.Listener)
		log.Fatal(http.ListenAndServe(c.Annotate.Listener, router))
	},
}

func init() {
	betaCmd.AddCommand(annotateCmd)
	annotateCmd.PersistentFlags().StringVarP(&annotateAuthHeaderName, "auth-header-name", "", "", "HTTP header name to check")
	annotateCmd.PersistentFlags().StringVarP(&annotateAuthHeaderValues, "auth-header-values", "", "", "HTTP header values to be allowed, comma separated list")
}
