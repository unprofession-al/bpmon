//go:generate esc -o static.go -pkg dashboard -prefix static static

package dashboard

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/configs"
	"github.com/unprofession-al/bpmon/periphery/webhelpers"
)

var bps bpmon.BusinessProcesses

func Setup(conf configs.DashboardConf, bpin bpmon.BusinessProcesses) (http.Handler, error) {
	bps = bpin
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/api/bps/", ListBPsHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}", GetBPHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}/kpis", ListKPIsHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}/kpis/{kpi}", GetKPIHandler).Methods("GET")

	if conf.Static == "" {
		assetHandler := webhelpers.GetAssetHandler("/assets/")
		r.PathPrefix("/assets/").Handler(assetHandler)

		statikFS := FS(false)
		r.PathPrefix("/").Handler(http.FileServer(statikFS))
	} else {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.Static)))
	}

	chain := alice.New().Then(r)

	return chain, nil
}
