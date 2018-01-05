package annotate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	wh "github.com/unprofession-al/bpmon/periphery/webhelpers"
	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

func ListEvents(res http.ResponseWriter, req *http.Request) {
	start, end := wh.GetStartEnd(req)

	// TODO: This two values should not be hard coded...
	interval, _ := time.ParseDuration("300s")
	stati := []status.Status{status.StatusNOK}

	out, err := pp.GetEvents(store.KindBusinessProcess, start, end, interval, stati)
	if err != nil {
		msg := fmt.Sprintf("An error occured: %s", err.Error())
		wh.Respond(res, req, http.StatusInternalServerError, msg)
		return
	}

	wh.Respond(res, req, http.StatusOK, out)
}

func AnnotateEvent(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	id := store.ID(vars["event"])

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		wh.Respond(res, req, http.StatusInternalServerError, err.Error())
		return
	}
	message := string(b)

	out, err := pp.AnnotateEvent(id, message)
	if err != nil {
		wh.Respond(res, req, http.StatusInternalServerError, err.Error())
		return
	}

	wh.Respond(res, req, http.StatusCreated, out)
}
