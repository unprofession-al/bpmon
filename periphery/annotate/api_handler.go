package annotate

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	wh "github.com/unprofession-al/bpmon/periphery/webhelpers"
	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

func ListEvents(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	kind, ok := vars["kind"]
	if !ok {
		wh.Respond(res, req, http.StatusNotFound, errors.New("type not found"))
		return
	}

	start, end := wh.GetStartEnd(req)
	interval, _ := time.ParseDuration("300s")
	stati := []status.Status{status.Nok}

	var out []store.Event
	for _, bp := range bps {
		rs := store.ResultSet{
			Tags: map[string]string{store.IdentifierBusinessProcess: bp.Id},
		}
		if strings.ToUpper(kind) == store.IdentifierBusinessProcess {
			events, err := pp.GetEvents(rs, start, end, interval, stati)
			if err != nil {
				wh.Respond(res, req, http.StatusInternalServerError, err.Error())
				return
			}
			out = append(out, events...)
		}
	}

	wh.Respond(res, req, http.StatusOK, out)
}

func AnnotateEvent(res http.ResponseWriter, req *http.Request) {}
