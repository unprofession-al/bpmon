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
var pp store.Accessor

var routes = make(map[string]wh.Leafs)

func Setup(conf configs.DashboardConf, bpsIn bpmon.BusinessProcesses, ppIn store.Accessor, tokenAuth bool, recipientHashes map[string]string, recipientsHeaderAuth bool, recipientsHeaderName string) (http.Handler, error) {
	pp = ppIn
	bps = bpsIn

	r := mux.NewRouter().StrictSlash(true)

	apiRouter := mux.NewRouter()
	api := apiRouter.PathPrefix("/api/").Subrouter()
	wh.PopulateRouter(api, routes)
	if tokenAuth {
		ta := wh.TokenAuth{Tokens: recipientHashes}
		r.Handle("/api/{_:.*}", ta.Create(apiRouter))
	} else if recipientsHeaderAuth {
		r.Handle("/api/{_:.*}", wh.RecipientsHeaderAuth(apiRouter, recipientsHeaderName))
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
