package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/unprofession-al/bpmon/icinga"
)

func MockIcingaServicesHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	env, ok := vars["env"]
	if !ok {
		Respond(res, req, http.StatusNotFound, "No environment passed")
		return
	}

	all := false
	var host, svc string
	var err error

	hostSvcPair := req.URL.Query()["service"]
	if len(hostSvcPair) < 1 {
		all = true
	} else {
		host, svc, err = splitHostServicePair(hostSvcPair)
		if err != nil {
			Respond(res, req, http.StatusNotFound, "No environment passed")
			return
		}
	}

	t := icinga.Timestamp{time.Now()}

	data := icinga.IcingaStatusResponse{}
	if all {
		data, err = envs.ToIcinga(env, t)
		if err != nil {
			Respond(res, req, http.StatusNotFound, "Environment not found ")
			return
		}
	} else {
		data, err = envs.SingleToIcinga(env, host, svc, t)
		if err != nil {
			Respond(res, req, http.StatusNotFound, "Environment not found ")
			return
		}
	}
	Respond(res, req, http.StatusOK, data)
}

func MockIcingaAcknowledgeHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	env, ok := vars["env"]
	if !ok {
		Respond(res, req, http.StatusNotFound, "No environment passed")
		return
	}

	host, svc, err := splitHostServicePair(req.URL.Query()["service"])
	if err != nil {
		Respond(res, req, http.StatusInternalServerError, err.Error())
		return
	}

	e, err := envs.Get(env)
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	h, err := e.Get(host)
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	s, err := h.Get(svc)
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	s.Acknowledgement = true
	Respond(res, req, http.StatusOK, "ok")
}

func splitHostServicePair(hostSvcPair []string) (string, string, error) {
	if len(hostSvcPair) == 1 {
		tokens := strings.SplitN(hostSvcPair[0], "!", 2)
		if len(tokens) != 2 {
			return "", "", fmt.Errorf("URL params `%s` unknown", hostSvcPair[0])
		}
		return tokens[0], tokens[1], nil

	} else if len(hostSvcPair) > 1 {
		return "", "", fmt.Errorf("To many URL params `serivce` found: %s", hostSvcPair)
	}
	return "", "", fmt.Errorf("No URL param `service` found")
}
