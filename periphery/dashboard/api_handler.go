package dashboard

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/unprofession-al/bpmon"
	wh "github.com/unprofession-al/bpmon/periphery/webhelpers"
	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

func ListBPsHandler(res http.ResponseWriter, req *http.Request) {
	list := make(map[string]string)

	if recipient := req.Context().Value(wh.KeyRecipient); recipient != nil {
		for _, bp := range bps.GetByRecipient(recipient.(string)) {
			list[bp.ID] = bp.Name
		}
	} else {
		for _, bp := range bps {
			list[bp.ID] = bp.Name
		}
	}

	wh.Respond(res, req, http.StatusOK, list)
}

func GetBPTimelineHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bpid := vars["bp"]

	// name := ""
	found := false
	for _, bp := range bps {
		if bp.ID == bpid {
			found = true
			//name = bp.Name
		}
	}

	if !found {
		msg := fmt.Sprintf("Business process %s not found", bpid)
		wh.Respond(res, req, http.StatusNotFound, msg)
		return
	}

	start, end := wh.GetStartEnd(req)

	where := store.ResultSet{
		Tags: map[string]string{"BP": bpid},
	}
	interval, _ := time.ParseDuration("300s")
	points, err := pp.GetSpans(where, start, end, interval, []status.Status{})
	if err != nil {
		msg := fmt.Sprintf("An error occured: %s", err.Error())
		wh.Respond(res, req, http.StatusInternalServerError, msg)
		return
	}

	wh.Respond(res, req, http.StatusOK, points)
}

func ListKPIsHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bpid := vars["bp"]

	list := make(map[string]string)
	found := false
	for _, bp := range bps {
		if bp.ID == bpid {
			found = true
			for _, kpi := range bp.Kpis {
				list[kpi.ID] = kpi.Name
			}
		}
	}
	if !found {
		msg := fmt.Sprintf("Business process %s not found", bpid)
		wh.Respond(res, req, http.StatusNotFound, msg)
		return
	}

	wh.Respond(res, req, http.StatusOK, list)
}

func GetKPITimelineHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bpid := vars["bp"]
	kpiid := vars["kpi"]

	bp := bpmon.BP{}
	found := false
	for _, currentBP := range bps {
		if currentBP.ID == bpid {
			found = true
			bp = currentBP
		}
	}

	if !found {
		msg := fmt.Sprintf("Business process %s not found", bpid)
		wh.Respond(res, req, http.StatusNotFound, msg)
		return
	}

	found = false
	for _, currentKPI := range bp.Kpis {
		if currentKPI.ID == kpiid {
			found = true
		}
	}

	if !found {
		msg := fmt.Sprintf("KPI %s of Business process %s not found", kpiid, bpid)
		wh.Respond(res, req, http.StatusNotFound, msg)
		return
	}

	start, end := wh.GetStartEnd(req)

	where := store.ResultSet{
		Tags: map[string]string{"BP": bpid, "KPI": kpiid},
	}

	interval, _ := time.ParseDuration("300s")
	points, err := pp.GetSpans(where, start, end, interval, []status.Status{})
	if err != nil {
		msg := fmt.Sprintf("An error occured: %s", err.Error())
		wh.Respond(res, req, http.StatusInternalServerError, msg)
		return
	}

	wh.Respond(res, req, http.StatusOK, points)
}
