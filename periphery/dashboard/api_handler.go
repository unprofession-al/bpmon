package dashboard

import (
	"net/http"

	wh "github.com/unprofession-al/bpmon/periphery/webhelpers"
)

func ListBPsHandler(res http.ResponseWriter, req *http.Request) {
	out := "bla"
	wh.Respond(res, req, http.StatusOK, out)
}

func GetBPHandler(res http.ResponseWriter, req *http.Request) {
	out := "bla"
	wh.Respond(res, req, http.StatusOK, out)
}

func ListKPIsHandler(res http.ResponseWriter, req *http.Request) {
	out := "bla"
	wh.Respond(res, req, http.StatusOK, out)
}

func GetKPIHandler(res http.ResponseWriter, req *http.Request) {
	out := "bla"
	wh.Respond(res, req, http.StatusOK, out)
}
