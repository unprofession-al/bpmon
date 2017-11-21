package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func ServiceHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	env, ok := vars["env"]
	if !ok {
		Respond(res, req, http.StatusNotFound, "No environment passed")
		return
	}

	hostSvcPair := req.URL.Query()["service"]

	all := true
	host := ""
	svc := ""
	if len(hostSvcPair) == 1 {
		tokens := strings.SplitN(hostSvcPair[0], "!", 2)
		if len(tokens) != 2 {
			out := fmt.Sprintf("URL params `%s` unknown", req.URL.Query())
			Respond(res, req, http.StatusInternalServerError, out)
			return
		}
		host = tokens[0]
		svc = tokens[1]
		all = false
	} else if len(hostSvcPair) > 1 {
		out := fmt.Sprintf("To many URL params `serivce` found: %s", hostSvcPair)
		Respond(res, req, http.StatusNotFound, out)
		return
	}

	t := Timestamp{time.Now()}

	data := IcingaStatusResponse{}
	var err error
	if all {
		data, err = envs.ToIcinga(env, t)
		if err != nil {
			Respond(res, req, http.StatusNotFound, "Environment not found ")
			return
		}
	} else {
		data, err = envs.SingleToIcinga(env, host, svc)
		if err != nil {
			Respond(res, req, http.StatusNotFound, "Environment not found ")
			return
		}
	}
	Respond(res, req, http.StatusOK, data)
}
