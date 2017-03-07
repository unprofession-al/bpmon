package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
)

var (
	printTimestamps bool
	printValues     bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all business process checks and print to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		c, b, err := bpmon.Configure(cfgFile, cfgSection, bpPath, bpPattern)
		if err != nil {
			log.Fatal(err)
		}

		i, err := bpmon.NewIcinga(c.Icinga, c.Rules)
		if err != nil {
			log.Fatal(err)
		}
		for _, bp := range b {
			rs := bp.Status(i)
			fmt.Println(rs.PrettyPrint(0, printTimestamps, printValues))
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().BoolVarP(&printTimestamps, "ts", "t", false, "print timestamps of measurement")
	runCmd.PersistentFlags().BoolVarP(&printValues, "vals", "v", false, "print raw  measurement results if available")
}
