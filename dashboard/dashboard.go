package dashboard

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/store"
)

type Dashboard struct {
	bp       bpmon.BusinessProcesses
	store    store.Accessor
	listener string
	handler  http.Handler
}

const (
	KeyRecipients key = iota
)

func New(c Config, bp bpmon.BusinessProcesses, store store.Accessor, tokenAuth bool, recipientHashes map[string]string, recipientsHeaderAuth bool, recipientsHeaderName string) Dashboard {
	d := Dashboard{
		bp:       bp,
		listener: c.Listener,
		store:    store,
	}

	r := mux.NewRouter().StrictSlash(true)

	apiRouter := mux.NewRouter()
	api := apiRouter.PathPrefix("/api/").Subrouter()
	PopulateRouter(api, d.getRoutes())
	if tokenAuth {
		m := TokenAuth{
			Tokens:     recipientHashes,
			ContextKey: KeyRecipients,
		}
		r.Handle("/api/{_:.*}", m.Inject(apiRouter))
	} else if recipientsHeaderAuth {
		m := HeaderAuth{
			HeaderName: recipientsHeaderName,
			ContextKey: KeyRecipients,
		}
		r.Handle("/api/{_:.*}", m.Inject(apiRouter))
	} else {
		r.Handle("/api/{_:.*}", apiRouter)
	}

	if c.Static != "" {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(c.Static)))
	}

	d.handler = alice.New().Then(r)

	return d
}

func (d Dashboard) Run() {
	fmt.Printf("Serving Dashboard at http://%s\nPress CTRL-c to stop...\n", d.listener)
	log.Fatal(http.ListenAndServe(d.listener, d.handler))
}

func (d Dashboard) getRoutes() map[string]Leafs {
	return map[string]Leafs{
		"v1": Leafs{
			"annotate": Leaf{
				L: Leafs{
					"{id}": Leaf{
						E: Endpoints{
							"POST": Endpoint{N: "Annotate", H: d.AnnotateHandler},
						},
					},
				},
			},
			"bps": Leaf{
				E: Endpoints{
					"GET": Endpoint{N: "ListBPs", H: d.ListBPsHandler},
				},
				L: Leafs{
					"{bp}": Leaf{
						E: Endpoints{
							"GET": Endpoint{N: "GetBPSpans", H: d.GetBPTimelineHandler},
						},
						L: Leafs{
							"kpis": Leaf{
								E: Endpoints{
									"GET": Endpoint{N: "ListKPIs", H: d.ListKPIsHandler},
								},
								L: Leafs{
									"{kpi}": Leaf{
										E: Endpoints{
											"GET": Endpoint{N: "GetKPISpans", H: d.GetKPITimelineHandler},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
