package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/checker"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var (
	printTimestamps  bool
	printValues      bool
	printResponsible bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all business process checks and print to stdout",
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

		s, err := c.Section(cfgSection)
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		i, err := checker.New(s.Checker)
		if err != nil {
			log.Fatal(err)
		}

		r := i.DefaultRules()
		err = r.Merge(s.Rules)
		if err != nil {
			log.Fatal(err)
		}

		a, err := s.Availabilities.Parse()
		if err != nil {
			log.Fatal(err)
		}

		b, err := bpmon.LoadBP(bpPath, bpPattern, a, s.GlobalRecipient)
		if err != nil {
			log.Fatal(err)
		}

		p, _ := store.New(s.Store)
		for _, bp := range b {
			rs := bp.Status(i, p, r)
			if s.Store.GetLastStatus {
				rs.AddPreviousStatus(p, s.Store.SaveOK)
			}
			fmt.Println(rs.PrettyPrint(0, printTimestamps, printValues, printResponsible))
		}

	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().BoolVarP(&printTimestamps, "ts", "t", false, "print timestamps of measurement")
	runCmd.PersistentFlags().BoolVarP(&printValues, "measurements", "m", false, "print raw measurement results if available")
	runCmd.PersistentFlags().BoolVarP(&printResponsible, "resp", "r", false, "print responsible of measurement")
}
