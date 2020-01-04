package dashboard

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/unprofession-al/bpmon"
	"github.com/unprofession-al/bpmon/internal/store"
)

type Dashboard struct {
	bp         bpmon.BusinessProcesses
	store      store.Accessor
	listener   string
	handler    http.Handler
	grantWrite []string
	auth       bool
}

const (
	KeyRecipients key = iota
)

func New(c Config, bp bpmon.BusinessProcesses, store store.Accessor, authPepper string, authHeader string) (Dashboard, string, error) {
	msg := ""

	d := Dashboard{
		bp:         bp,
		listener:   c.Listener,
		store:      store,
		grantWrite: c.GrantWrite,
	}

	r := mux.NewRouter().StrictSlash(true)

	apiRouter := mux.NewRouter()
	api := apiRouter.PathPrefix("/api/").Subrouter()
	PopulateRouter(api, d.getRoutes())

	authorization := Authorization{
		RecipientContextKey: KeyRecipients,
		OnAuthErrorReturn:   http.StatusNotFound,
		BP:                  bp,
		ProtectPattern:      regexp.MustCompile("^/api/v1/bps/.+"),
	}

	if authPepper != "" && authHeader != "" {
		return d, msg, fmt.Errorf("ERROR: pepper and recipients-header are set, only one is allowed.")
	} else if authPepper == "" && authHeader == "" {
		d.auth = false
		msg = "WARNING: No pepper or recipients-header is provided, all information are accessible without auth..."
		r.Handle("/api/{_:.*}", apiRouter)
	} else if authHeader != "" {
		d.auth = true
		msg = fmt.Sprintf("Recipients-header is provided, using HTTP Header '%s' to read recipients...\n", authHeader)
		m := HeaderAuth{
			HeaderName: authHeader,
			ContextKey: KeyRecipients,
		}
		r.Handle("/api/{_:.*}", alice.New(m.Wrap, authorization.Wrap).Then(apiRouter))
	} else if authPepper != "" {
		d.auth = true
		var recipientHashes map[string]string
		msg = fmt.Sprintf("Pepper is provided, generating auth hashes...\n")
		recipientHashes = bp.GenerateRecipientHashes(authPepper)
		msg = msg + fmt.Sprintf("%15s: %s\n", "Recipient", "Hash")
		for k, v := range recipientHashes {
			msg = msg + fmt.Sprintf("%15s: %s\n", v, k)
		}
		m := TokenAuth{
			Tokens:     recipientHashes,
			Param:      "authtoken",
			ContextKey: KeyRecipients,
		}
		r.Handle("/api/{_:.*}", alice.New(m.Wrap, authorization.Wrap).Then(apiRouter))
	}

	if c.Static != "" {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(c.Static)))
	}

	d.handler = alice.New().Then(r)

	return d, msg, nil
}

func (d Dashboard) Run() {
	fmt.Printf("Serving Dashboard at http://%s\nPress CTRL-c to stop...\n", d.listener)
	log.Fatal(http.ListenAndServe(d.listener, d.handler))
}

func (d Dashboard) getRoutes() map[string]Leafs {
	return map[string]Leafs{
		"v1": {
			"whoami": Leaf{
				E: Endpoints{
					"GET": Endpoint{N: "WhoAmI", H: d.WhoamiHandler},
				},
			},
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
