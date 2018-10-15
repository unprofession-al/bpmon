package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/spf13/cobra"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Run all business process checks and trigger a custom command if issue occure",
	Run: func(cmd *cobra.Command, args []string) {
		c, _, err := config.New(cfgFile, injectDefaults)
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

		t := template.Must(template.New("t1").Parse(s.Trigger.Template))

		filterBy := []status.Status{status.StatusNOK}
		var sets []store.ResultSet
		for _, bp := range b {
			rs := bp.Status(i, nil, r)
			if s.Store.GetLastStatus {
				rs.AddPreviousStatus(p, s.Store.SaveOK)
			}
			set, stripped := rs.FilterByStatus(filterBy)
			if !stripped {
				sets = append(sets, set)
			}
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
	RootCmd.AddCommand(triggerCmd)
}
