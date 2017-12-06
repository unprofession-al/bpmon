//go:generate statik -src=./static -f

package dashboard

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rakyll/statik/fs"

	_ "github.com/unprofession-al/bpmon/periphery/dashboard/statik"
	"github.com/unprofession-al/bpmon/periphery/webhelpers"
)

func Router(conf Conf) (http.Handler, error) {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/api/bps/", ListBPsHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}", GetBPHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}/kpis", ListKPIsHandler).Methods("GET")
	r.HandleFunc("/api/bps/{bp}/kpis/{kpi}", GetKPIHandler).Methods("GET")

	if conf.Static == "" {
		assetHandler := webhelpers.GetAssetHandler("/assets/")

		//assetHandler := webhelpers.GetDummyHandler("/assets/")
		r.PathPrefix("/assets/").Handler(assetHandler)

		statikFS, err := fs.New()
		if err != nil {
			return nil, err
		}
		r.PathPrefix("/").Handler(http.FileServer(statikFS))
	} else {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.Static)))
	}

	chain := alice.New().Then(r)

	return chain, nil
}
