package dashboard

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/store"
)

var bps bpmon.BusinessProcesses
var pp store.Accessor

var routes = make(map[string]Leafs)

func Setup(conf Config, bpsIn bpmon.BusinessProcesses, ppIn store.Accessor, tokenAuth bool, recipientHashes map[string]string, recipientsHeaderAuth bool, recipientsHeaderName string) (http.Handler, error) {
	pp = ppIn
	bps = bpsIn

	r := mux.NewRouter().StrictSlash(true)

	apiRouter := mux.NewRouter()
	api := apiRouter.PathPrefix("/api/").Subrouter()
	PopulateRouter(api, routes)
	if tokenAuth {
		ta := TokenAuth{Tokens: recipientHashes}
		r.Handle("/api/{_:.*}", ta.Create(apiRouter))
	} else if recipientsHeaderAuth {
		r.Handle("/api/{_:.*}", RecipientsHeaderAuth(apiRouter, recipientsHeaderName))
	} else {
		r.Handle("/api/{_:.*}", apiRouter)
	}

	if conf.Static != "" {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.Static)))
	}

	chain := alice.New().Then(r)

	return chain, nil
}
