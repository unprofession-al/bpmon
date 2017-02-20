package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Insert data into influx db",
	Run: func(cmd *cobra.Command, args []string) {
		c, b, err := configure()
		if err != nil {
			log.Fatal(err)
		}

		i, err := NewIcinga(c.Icinga)
		if err != nil {
			log.Fatal(err)
		}
		infl, _ := NewInflux(c.Influx)
		for _, bp := range b {
			log.Println("Processing " + bp.Name)
			rs := bp.Status(i)
			err = infl.Write(rs)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(writeCmd)
}
