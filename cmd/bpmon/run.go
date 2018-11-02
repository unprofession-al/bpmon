package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/runners"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var (
	runParams []string
	runList   bool
	runAdHoc  string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all business process checks and render the results using a given template",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runnerName := "default"
		if len(args) > 0 {
			runnerName = args[0]
		}

		cfg := fmt.Sprintf("%s/%s", cfgBase, cfgFile)
		c, _, err := config.NewFromFile(cfg, injectDefaults)
		if err != nil {
			log.Fatal(err)
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

		runnerDir := fmt.Sprintf("%s/%s", cfgBase, s.Env.Runners)
		run, err := runners.New(runnerDir)
		if err != nil {
			log.Fatal(err)
		}

		if runList {
			for name, runner := range run {
				p := ""
				for key, desc := range runner.Parameters {
					p += fmt.Sprintf("\t%s: %s\n", key, desc)
				}
				fmt.Printf("%s\n\t%s\n%s", name, runner.Description, p)
			}
			os.Exit(0)
		}

		if runAdHoc != "" {
			runnerName, err = run.AdHoc(runAdHoc)
			if err != nil {
				log.Fatal(err)
			}
		}

		runner, ok := run[runnerName]
		if !ok {
			msg := fmt.Sprintf("Template '%s' not found in section '%s'", runnerName, cfgSection)
			log.Fatal(msg)
		}

		params := map[string]string{}

		for _, param := range runParams {
			tokens := strings.SplitN(param, "=", 2)
			if len(tokens) == 2 {
				params[tokens[0]] = tokens[1]
			} else {
				fmt.Fprintf(os.Stderr, "Param '%s' is invalid, must match pattern [key]=[value] ", param)
			}
		}

		if err != nil {
			msg := fmt.Sprintf("Could not parse params passed as JSON: %s", err.Error())
			log.Fatal(msg)
		}

		for k, v := range runner.Parameters {
			if _, ok := params[k]; ok {
				continue
			}
			fmt.Fprintf(os.Stderr, "Enter '%s' (%s): ", k, v)
			reader := bufio.NewReader(os.Stdin)
			params[k], err = reader.ReadString('\n')
			params[k] = strings.TrimSuffix(params[k], "\n")
			if err != nil {
				msg := fmt.Sprintf("Could not read input for parameter '%s':  %s", k, err.Error())
				log.Fatal(msg)
			}
		}

		var sets []store.ResultSet
		for _, bp := range b {
			rs := bp.Status(i, nil, r)
			if s.Store.GetLastStatus {
				rs.AddPreviousStatus(p, s.Store.SaveOK)
			}
			sets = append(sets, rs)
			if runner.ForEach {
				data := struct {
					BP         []store.ResultSet
					Config     config.ConfigSection
					Parameters map[string]string
				}{
					BP:         []store.ResultSet{rs},
					Config:     s,
					Parameters: params,
				}
				err = runner.Exec(data)
				if err != nil {
					break
				}
			}
		}

		if !runner.ForEach {
			data := struct {
				BP         []store.ResultSet
				Config     config.ConfigSection
				Parameters map[string]string
			}{
				BP:         sets,
				Config:     s,
				Parameters: params,
			}
			err = runner.Exec(data)
		}

		if err != nil {
			msg := fmt.Sprintf("Error while Executing runner: %s", err.Error())
			log.Fatal(msg)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringSliceVar(&runParams, "params", []string{}, "Provide template parameters")
	runCmd.PersistentFlags().BoolVar(&runList, "list", false, "print a list of available runners")
	runCmd.PersistentFlags().StringVar(&runAdHoc, "adhoc", "", "pass a runner template as param")
}
