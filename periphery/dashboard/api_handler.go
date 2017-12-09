package dashboard

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/unprofession-al/bpmon"
	wh "github.com/unprofession-al/bpmon/periphery/webhelpers"
)

func ListBPsHandler(res http.ResponseWriter, req *http.Request) {
	list := make(map[string]string)
	for _, bp := range bps {
		list[bp.Id] = bp.Name
	}
	wh.Respond(res, req, http.StatusOK, list)
}

func GetBPTimelineHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bpid := vars["bp"]

	// name := ""
	found := false
	for _, bp := range bps {
		if bp.Id == bpid {
			found = true
			//name = bp.Name
		}
	}

	if !found {
		msg := fmt.Sprintf("Business process %s not found", bpid)
		wh.Respond(res, req, http.StatusNotFound, msg)
		return
	}

	end := time.Now()
	start := end.AddDate(0, -1, 0)

	where := map[string]string{
		"BP":      fmt.Sprintf("'%s'", bpid),
		"changed": "true",
	}
	points, err := influx.GetEvents("BP", where, start, end)
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
		if bp.Id == bpid {
			found = true
			for _, kpi := range bp.Kpis {
				list[kpi.Id] = kpi.Name
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
		if currentBP.Id == bpid {
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
		if currentKPI.Id == kpiid {
			found = true
		}
	}

	if !found {
		msg := fmt.Sprintf("KPI %s of Business process %s not found", kpiid, bpid)
		wh.Respond(res, req, http.StatusNotFound, msg)
		return
	}

	end := time.Now()
	start := end.AddDate(0, -1, 0)

	where := map[string]string{
		"BP":  fmt.Sprintf("'%s'", bpid),
		"KPI": fmt.Sprintf("'%s'", kpiid),
	}
	points, err := influx.GetEvents("KPI", where, start, end)
	if err != nil {
		msg := fmt.Sprintf("An error occured: %s", err.Error())
		wh.Respond(res, req, http.StatusInternalServerError, msg)
		return
	}

	wh.Respond(res, req, http.StatusOK, points)
}
