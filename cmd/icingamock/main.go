package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var (
	configFile string
	envDir     string
	cfg        Configuration
	envs       Environments
)

func init() {
	flag.StringVar(&configFile, "conf", "./config/icingamock.yaml", "path to config file")
	flag.StringVar(&envDir, "env", "./config/env.d/", "environment setup files")
}

func main() {
	flag.Parse()
	var err error
	cfg, err = Configure(configFile)
	if err != nil {
		log.Fatal(err)
	}

	envs, err := LoadEnvs(envDir, "*.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(envs)

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/v1/objects/services", ServiceHandler)

	chain := alice.New().Then(r)

	log.Fatal(http.ListenAndServe(cfg.Listener.Address+":"+cfg.Listener.Port, chain))
}

// /objects/services?service=%s!%s
