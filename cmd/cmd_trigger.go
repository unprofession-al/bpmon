package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/spf13/cobra"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Run all business process checks and trigger temploted command on BP issues",
	Run: func(cmd *cobra.Command, args []string) {
		c, b, err := configure()
		if err != nil {
			log.Fatal(err)
		}

		t := template.Must(template.New("t1").Parse(c.Trigger.Template))

		i, err := NewIcinga(c.Icinga)
		if err != nil {
			log.Fatal(err)
		}
		stripBy := []Status{StatusUnknown, StatusOK}
		var sets []ResultSet
		for _, bp := range b {
			rs := bp.Status(i)
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
