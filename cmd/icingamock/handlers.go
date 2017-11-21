package main

import (
	"fmt"
	"net/http"
	"strings"
)

func ServiceHandler(res http.ResponseWriter, req *http.Request) {
	out := ""
	hostSvcPair := req.URL.Query()["service"]

	all := true
	host := ""
	svc := ""
	if len(hostSvcPair) == 1 {
		tokens := strings.SplitN(hostSvcPair[0], "!", 2)
		if len(tokens) != 2 {
			out = fmt.Sprintf("URL params `%s` unknown", req.URL.Query())
			Respond(res, req, http.StatusInternalServerError, out)
			return
		}
		host = tokens[0]
		svc = tokens[1]
		all = false
	} else if len(hostSvcPair) > 1 {
		out = fmt.Sprintf("To many URL params `serivce` found: %s", hostSvcPair)
		Respond(res, req, http.StatusNotFound, out)
		return
	}

	if all {
		out = fmt.Sprintf("All Hosts/Services")
	} else {
		out = fmt.Sprintf("Host: %s / Service: %s", host, svc)
	}
	Respond(res, req, http.StatusOK, out)
}
