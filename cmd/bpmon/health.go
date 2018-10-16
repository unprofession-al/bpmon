package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/health"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health of store and checker",
	Run: func(cmd *cobra.Command, args []string) {
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

		s, i, _, _, p, err := fromSection(c, cfgSection)
		if err != nil {
			msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", cfgSection, cfgFile, err.Error())
			log.Fatal(msg)
		}

		h, err := health.New(s.Health)
		if err != nil {
			log.Fatal(err)
		}

		rs := h.Check(i, p)

		fmt.Println(rs.PrettyPrint(0, true, true, true))
	},
}

func init() {
	RootCmd.AddCommand(healthCmd)
}
