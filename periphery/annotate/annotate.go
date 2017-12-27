//go:generate esc -o static.go -pkg annotate -prefix static static

package annotate

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/configs"
	wh "github.com/unprofession-al/bpmon/periphery/webhelpers"
	"github.com/unprofession-al/bpmon/store"
)

var bps bpmon.BusinessProcesses
var pp store.Store
var routes = make(map[string]wh.Leafs)

func Setup(conf configs.AnnotateConf, bpin bpmon.BusinessProcesses, ppin store.Store) (http.Handler, error) {
	pp = ppin
	bps = bpin

	r := mux.NewRouter().StrictSlash(true)
	api := r.PathPrefix("/api/").Subrouter()
	wh.PopulateRouter(api, routes)

	if conf.Static == "" {
		assetHandler := wh.GetAssetHandler("/assets/")
		r.PathPrefix("/assets/").Handler(assetHandler)

		r.PathPrefix("/").Handler(http.FileServer(FS(false)))
	} else {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.Static)))
	}

	chain := alice.New().Then(r)

	return chain, nil
}
