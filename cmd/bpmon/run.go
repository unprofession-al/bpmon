package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/spf13/cobra"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all business process checks and render the results using a given template",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := "default"
		if len(args) > 0 {
			templateName = args[0]
		}

		c, _, err := config.NewFromFile(cfgFile, injectDefaults)
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

		t := template.Must(template.New("t1").Parse(s.Templates[templateName]))

		var sets []store.ResultSet
		for _, bp := range b {
			rs := bp.Status(i, nil, r)
			if s.Store.GetLastStatus {
				rs.AddPreviousStatus(p, s.Store.SaveOK)
			}
			sets = append(sets, rs)
		}
		var command bytes.Buffer
		err = t.Execute(&command, sets)
		if err != nil {
			log.Fatal(err)
		}

		if len(sets) > 0 {
			fmt.Println(command.String())
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
