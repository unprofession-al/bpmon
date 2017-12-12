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
var ep bpmon.EventProvider

func Setup(conf configs.DashboardConf, bpin bpmon.BusinessProcesses, epin bpmon.EventProvider) (http.Handler, error) {
	ep = epin
	bps = bpin
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/api/bps/", ListBPsHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}", GetBPTimelineHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}/kpis", ListKPIsHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}/kpis/{kpi}", GetKPITimelineHandler).Methods("GET")

	if conf.Static == "" {
		assetHandler := webhelpers.GetAssetHandler("/assets/")
		r.PathPrefix("/assets/").Handler(assetHandler)

		r.PathPrefix("/").Handler(http.FileServer(FS(false)))
	} else {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.Static)))
	}

	chain := alice.New().Then(r)

	return chain, nil
}
