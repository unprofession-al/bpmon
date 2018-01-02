//go:generate esc -o static.go -pkg dashboard -prefix static static

package dashboard

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

func Setup(conf configs.DashboardConf, bpsIn bpmon.BusinessProcesses, ppIn store.Store, auth bool, recipientHashes map[string]string) (http.Handler, error) {
	pp = ppIn
	bps = bpsIn

	r := mux.NewRouter().StrictSlash(true)

	apiRouter := mux.NewRouter()
	api := apiRouter.PathPrefix("/api/").Subrouter()
	wh.PopulateRouter(api, routes)
	if auth {
		ta := wh.TokenAuth{Tokens: recipientHashes}
		r.Handle("/api/{_:.*}", ta.Create(apiRouter))
	} else {
		r.Handle("/api/{_:.*}", apiRouter)
	}

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
