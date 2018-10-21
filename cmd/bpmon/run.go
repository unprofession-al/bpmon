package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	_ "github.com/unprofession-al/bpmon/checker/icinga"
	"github.com/unprofession-al/bpmon/config"
	"github.com/unprofession-al/bpmon/store"
	_ "github.com/unprofession-al/bpmon/store/influx"
)

var runParams []string

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

		templateData, ok := s.Templates[templateName]
		if !ok {
			msg := fmt.Sprintf("Template '%s' not found in section '%s'", templateName, cfgSection)
			log.Fatal(msg)
		}

		t, err := template.New(templateName).Parse(templateData.Template)
		if err != nil {
			msg := fmt.Sprintf("Could not parse template '%s' from section '%s':  %s", templateName, cfgSection, err.Error())
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

		for k, v := range templateData.Parameters {
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
		}

		data := struct {
			BP         []store.ResultSet
			Config     config.ConfigSection
			Parameters map[string]string
		}{
			BP:         sets,
			Config:     s,
			Parameters: params,
		}

		var command bytes.Buffer
		err = t.Execute(&command, data)
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
	runCmd.PersistentFlags().StringSliceVar(&runParams, "params", []string{}, "Provide template parameters")
}
