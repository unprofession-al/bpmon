package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/checker"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Run all business process checks and trigger a custom command if issue occure",
	Run: func(cmd *cobra.Command, args []string) {
		c, b, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			msg := fmt.Sprintf("Could not read section %s form file %s, error was %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		t := template.Must(template.New("t1").Parse(c.Trigger.Template))

		i, err := checker.New(c.Checker)
		if err != nil {
			log.Fatal(err)
		}

		r := i.DefaultRules()
		err = r.Merge(c.Rules)
		if err != nil {
			log.Fatal(err)
		}

		p, _ := store.New(c.Store)
		filterBy := []status.Status{status.NOK}
		var sets []store.ResultSet
		for _, bp := range b {
			rs := bp.Status(i, nil, r)
			if c.Store.GetLastStatus {
				rs.AddPreviousStatus(p, c.Store.SaveOK)
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
