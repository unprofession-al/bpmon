package dashboard

import (
	"net/http"

	"github.com/gorilla/mux"
)

func notImplemented(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNotImplemented)
	out := "Function Not Yet Implemented\n"
	res.Write([]byte(out))
}

// PopulateRouter appends all defined routes to a given gorilla mux router.
func PopulateRouter(router *mux.Router, routes map[string]Leafs) {
	for version, leafs := range routes {
		api := router.PathPrefix("/" + version).Subrouter()

		for pattern, leaf := range leafs {
			appendLeaf(pattern, leaf, api)
		}
	}
}

func appendLeaf(p string, l Leaf, router *mux.Router) {
	for method, endpoint := range l.E {
		h := endpoint.H
		if h == nil {
			h = notImplemented
		}
		router.
			Methods(method).
			Path("/" + p).
			Name(endpoint.N).
			Handler(h)
	}
	for pattern, leaf := range l.L {
		appendLeaf(p+"/"+pattern, leaf, router)
	}
}

type Leafs map[string]Leaf

type Leaf struct {
	E Endpoints `json:"endpoints,omitempty"`
	L Leafs     `json:"leafs,omitempty"`
}

type Endpoints map[string]Endpoint

type Endpoint struct {
	N string `json:"name"`
	H http.HandlerFunc
}
