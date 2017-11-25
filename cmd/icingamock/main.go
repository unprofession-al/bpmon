//go:generate statik -src=./static

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rakyll/statik/fs"

	_ "github.com/unprofession-al/bpmon/cmd/icingamock/statik"
)

var (
	configFile string
	envDir     string
	bpDir      string
	staticDir  string
	cfg        Configuration
	envs       *Environments
)

func init() {
	flag.StringVar(&configFile, "conf", "./config/icingamock.yaml", "path to config file")
	flag.StringVar(&envDir, "env", "", "environment setup files")
	flag.StringVar(&bpDir, "bp", "", "bpmon bp files")
	flag.StringVar(&staticDir, "static", "", "static html served at http root")
}

func main() {
	flag.Parse()
	var err error
	cfg, err = Configure(configFile)
	if err != nil {
		log.Fatal(err)
	}

	envs, err = LoadEnvs(envDir, "*.yaml")
	if err != nil {
		log.Fatal(err)
	}

	(*envs)["_"], err = LoadEnvFromBP(bpDir, "*.yaml")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/icinga/{env}/v1/objects/services", MockIcingaServicesHandler).Methods("GET")
	r.HandleFunc("/icinga/{env}/v1/actions/acknowledge-problem", MockIcingaAcknowledgeHandler).Methods("POST")
	r.HandleFunc("/api/envs/", ListEnvsHandler).Methods("GET")
	r.HandleFunc("/api/envs/{env}", GetEnvHandler).Methods("GET")
	r.HandleFunc("/api/envs/{env}/hosts/", ListHostsHandler).Methods("GET")
	r.HandleFunc("/api/envs/{env}/hosts/{host}", GetHostHandler).Methods("GET")
	r.HandleFunc("/api/envs/{env}/hosts/{host}/services/", ListServicesHandler).Methods("GET")
	r.HandleFunc("/api/envs/{env}/hosts/{host}/services/{service}", GetServiceHandler).Methods("GET")
	r.HandleFunc("/api/envs/{env}/hosts/{host}/services/{service}", UpdateServiceHandler).Methods("POST")

	if staticDir == "" {
		statikFS, err := fs.New()
		if err != nil {
			log.Fatal(err)
		}
		r.PathPrefix("/").Handler(http.FileServer(statikFS))
	} else {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))
	}

	chain := alice.New().Then(r)

	log.Fatal(http.ListenAndServe(cfg.Listener.Address+":"+cfg.Listener.Port, chain))
}
