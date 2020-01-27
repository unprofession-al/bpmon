package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unprofession-al/bpmon/internal/config"
	"github.com/unprofession-al/bpmon/internal/dashboard"
	"github.com/unprofession-al/bpmon/internal/runners"
	"github.com/unprofession-al/bpmon/internal/store"
	"gopkg.in/yaml.v2"

	_ "github.com/unprofession-al/bpmon/internal/checker/icinga"
	_ "github.com/unprofession-al/bpmon/internal/store/influx"
)

const baseEnv = "BPMON_BASE"

type App struct {
	cfg struct {
		// root
		cfgFile        string
		cfgBase        string
		cfgSection     string
		bpPattern      string
		verbose        bool
		injectDefaults bool

		// config
		configInitComments bool

		// dashboard
		dashboardPepper string
		dashboardHeader string
		dashboardStatic string

		// run
		runParams []string
		runList   bool
		runAdHoc  string
	}

	// entry point
	Execute func() error
}

func NewApp() *App {
	a := &App{}

	base := os.Getenv(baseEnv)
	if base == "" {
		base = "."
	}

	// root
	rootCmd := &cobra.Command{
		Use:   "bpmon",
		Short: "Montior business processes composed of Icinga checks",
	}
	rootCmd.PersistentFlags().StringVarP(&a.cfg.cfgFile, "config", "c", "config.yaml", "name the configuration file")
	rootCmd.PersistentFlags().StringVarP(&a.cfg.cfgBase, "base", "b", base, fmt.Sprintf("path of the directory where the configuration is located (default can be set via $%s)", baseEnv))
	rootCmd.PersistentFlags().StringVarP(&a.cfg.cfgSection, "section", "s", "default", "name of the section to be read")
	rootCmd.PersistentFlags().StringVarP(&a.cfg.bpPattern, "pattern", "p", "*.yaml", "pattern of business process configuration files to process")
	rootCmd.PersistentFlags().BoolVarP(&a.cfg.verbose, "verbose", "v", true, "print log output")
	rootCmd.PersistentFlags().BoolVarP(&a.cfg.injectDefaults, "defaults", "d", true, "inject defaults in main config file")
	a.Execute = rootCmd.Execute

	// config
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Various tools to verify your BPMON config",
	}
	rootCmd.AddCommand(configCmd)

	// config print
	configPrintCmd := &cobra.Command{
		Use:   "print",
		Short: "Print the given configurantion section as interpreted by BPMON to stdout",
		Run:   a.configPrintCmd,
	}
	configCmd.AddCommand(configPrintCmd)

	// config init
	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Print a configurantion with default values and comments to stdout",
		Run:   a.configInitCmd,
	}
	configInitCmd.PersistentFlags().BoolVarP(&a.cfg.configInitComments, "comments", "", true, "print documentation comments")
	configCmd.AddCommand(configInitCmd)

	// config raw
	configRawCmd := &cobra.Command{
		Use:   "raw",
		Short: "Print the configuration with its injected defaults to stdout",
		Run:   a.configRawCmd,
	}
	configCmd.AddCommand(configRawCmd)

	// dashboard
	dashboardCmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Run Dashboard Web UI",
		Run:   a.dashboardCmd,
	}
	dashboardCmd.PersistentFlags().StringVarP(&a.cfg.dashboardPepper, "pepper", "", "", "Pepper used to generate auth token")
	dashboardCmd.PersistentFlags().StringVarP(&a.cfg.dashboardHeader, "header", "", "", "HTTP header name to read recipients from")
	dashboardCmd.PersistentFlags().StringVarP(&a.cfg.dashboardStatic, "static", "", "", "Path to custom html frontend")
	rootCmd.AddCommand(dashboardCmd)

	// run
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run all business process checks and render the results using a given template",
		Args:  cobra.MaximumNArgs(1),
		Run:   a.runCmd,
	}
	runCmd.PersistentFlags().StringSliceVar(&a.cfg.runParams, "params", []string{}, "Provide template parameters")
	runCmd.PersistentFlags().BoolVar(&a.cfg.runList, "list", false, "print a list of available runners")
	runCmd.PersistentFlags().StringVar(&a.cfg.runAdHoc, "adhoc", "", "pass a runner template as param")
	rootCmd.AddCommand(runCmd)

	// write
	writeCmd := &cobra.Command{
		Use:   "write",
		Short: "Insert data into InfluxDB",
		Run:   a.writeCmd,
	}
	rootCmd.AddCommand(writeCmd)

	// version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run:   a.versionCmd,
	}
	rootCmd.AddCommand(versionCmd)

	return a
}

func (a App) configPrintCmd(cmd *cobra.Command, args []string) {
	cfg := fmt.Sprintf("%s/%s", a.cfg.cfgBase, a.cfg.cfgFile)
	c, _, err := config.NewFromFile(cfg, a.cfg.injectDefaults)
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

	s, _, r, _, _, err := fromSection(c, a.cfg.cfgSection, a.cfg.cfgBase, a.cfg.bpPattern)
	if err != nil {
		msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", a.cfg.cfgSection, a.cfg.cfgFile, err.Error())
		log.Fatal(msg)
	}

	s.Rules = r

	var out []byte
	out, _ = yaml.Marshal(s)
	fmt.Println(string(out))
}

func (a App) configInitCmd(cmd *cobra.Command, args []string) {
	out := config.ExampleYAML(a.cfg.configInitComments)
	fmt.Println(string(out))
}

func (a App) configRawCmd(cmd *cobra.Command, args []string) {
	cfg := fmt.Sprintf("%s/%s", a.cfg.cfgBase, a.cfg.cfgFile)
	_, raw, err := config.NewFromFile(cfg, a.cfg.injectDefaults)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(raw))
}

func (a App) dashboardCmd(cmd *cobra.Command, args []string) {
	cfg := fmt.Sprintf("%s/%s", a.cfg.cfgBase, a.cfg.cfgFile)
	c, _, err := config.NewFromFile(cfg, a.cfg.injectDefaults)
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

	s, _, _, bp, store, err := fromSection(c, a.cfg.cfgSection, a.cfg.cfgBase, a.cfg.bpPattern)
	if err != nil {
		msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", a.cfg.cfgSection, a.cfg.cfgFile, err.Error())
		log.Fatal(msg)
	}

	if a.cfg.dashboardStatic != "" {
		s.Dashboard.Static = a.cfg.dashboardStatic
	}

	d, msg, err := dashboard.New(s.Dashboard, bp, store, a.cfg.dashboardPepper, a.cfg.dashboardHeader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(msg)

	d.Run()
}

func (a App) runCmd(cmd *cobra.Command, args []string) {
	runnerName := "default"
	if len(args) > 0 {
		runnerName = args[0]
	}

	cfg := fmt.Sprintf("%s/%s", a.cfg.cfgBase, a.cfg.cfgFile)
	c, _, err := config.NewFromFile(cfg, a.cfg.injectDefaults)
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

	s, i, r, b, p, err := fromSection(c, a.cfg.cfgSection, a.cfg.cfgBase, a.cfg.bpPattern)
	if err != nil {
		msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", a.cfg.cfgSection, a.cfg.cfgFile, err.Error())
		log.Fatal(msg)
	}

	runnerDir := fmt.Sprintf("%s/%s", a.cfg.cfgBase, s.Env.Runners)
	run, err := runners.New(runnerDir)
	if err != nil {
		log.Fatal(err)
	}

	if a.cfg.runList {
		for name, runner := range run {
			p := ""
			for key, desc := range runner.Parameters {
				p += fmt.Sprintf("\t%s: %s\n", key, desc)
			}
			fmt.Printf("%s\n\t%s\n%s", name, runner.Description, p)
		}
		os.Exit(0)
	}

	if a.cfg.runAdHoc != "" {
		runnerName, err = run.AdHoc(a.cfg.runAdHoc)
		if err != nil {
			log.Fatal(err)
		}
	}

	runner, ok := run[runnerName]
	if !ok {
		msg := fmt.Sprintf("Template '%s' not found in section '%s'", runnerName, a.cfg.cfgSection)
		log.Fatal(msg)
	}

	params := map[string]string{}

	for _, param := range a.cfg.runParams {
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
}

func (a App) writeCmd(cmd *cobra.Command, args []string) {
	cfg := fmt.Sprintf("%s/%s", a.cfg.cfgBase, a.cfg.cfgFile)
	c, _, err := config.NewFromFile(cfg, a.cfg.injectDefaults)
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

	s, i, r, b, p, err := fromSection(c, a.cfg.cfgSection, a.cfg.cfgBase, a.cfg.bpPattern)
	if err != nil {
		msg := fmt.Sprintf("Could not read section '%s' from file '%s':  %s", a.cfg.cfgSection, a.cfg.cfgFile, err.Error())
		log.Fatal(msg)
	}
	for _, bp := range b {
		if a.cfg.verbose {
			log.Println("Processing " + bp.Name)
		}
		rs := bp.Status(i, p, r)
		if s.Store.GetLastStatus {
			rs.AddPreviousStatus(p, s.Store.SaveOK)
		}
		err = p.Write(&rs)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (a App) versionCmd(cmd *cobra.Command, args []string) {
	fmt.Println(versionInfo())
}
