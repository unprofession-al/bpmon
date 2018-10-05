//go:generate esc -o static.go -pkg annotate -prefix static static

package annotate

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/config"
	wh "github.com/unprofession-al/bpmon/periphery/webhelpers"
	"github.com/unprofession-al/bpmon/store"
)

var bps bpmon.BusinessProcesses
var pp store.Accessor
var routes = make(map[string]wh.Leafs)

func Setup(conf config.AnnotateConfig, bpin bpmon.BusinessProcesses, ppin store.Accessor, authHeaderName string, authHeaderValues []string) (http.Handler, error) {
	pp = ppin
	bps = bpin

	r := mux.NewRouter().StrictSlash(true)

	apiRouter := mux.NewRouter()
	api := apiRouter.PathPrefix("/api/").Subrouter()
	wh.PopulateRouter(api, routes)
	if authHeaderName != "" {
		r.Handle("/api/{_:.*}", wh.HeaderAuthMatcher(apiRouter, authHeaderName, authHeaderValues))
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

	// chain := alice.New(wh.Logger).Then(r)
	chain := alice.New().Then(r)

	return chain, nil
}
