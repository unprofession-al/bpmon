package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/icinga"
	"github.com/unprofession-al/bpmon/status"
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

		i, err := icinga.NewIcinga(c.Icinga, c.Rules)
		if err != nil {
			log.Fatal(err)
		}

		r := i.DefaultRules()
		r.Merge(c.Rules)
		if err != nil {
			log.Fatal(err)
		}

		infl, _ := bpmon.NewInflux(c.Influx)
		stripBy := []status.Status{status.Unknown, status.Ok}
		var sets []bpmon.ResultSet
		for _, bp := range b {
			rs := bp.Status(i, nil, r)
			if c.Influx.GetLastStatus {
				rs.AddPreviousStatus(infl, c.Influx.SaveOK)
			}
			set, stripped := rs.StripByStatus(stripBy)
			if !stripped {
				sets = append(sets, set)
			}
		}
		var command bytes.Buffer
		t.Execute(&command, sets)
		if len(sets) > 0 {
			fmt.Println(command.String())
		}
	},
}

func init() {
	RootCmd.AddCommand(triggerCmd)
}
