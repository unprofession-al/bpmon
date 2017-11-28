package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func ListEnvsHandler(res http.ResponseWriter, req *http.Request) {
	out := envs.List()
	Respond(res, req, http.StatusOK, out)
}

func GetEnvHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	env, err := envs.Get(vars["env"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	Respond(res, req, http.StatusOK, env)
}

func ListHostsHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	env, err := envs.Get(vars["env"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	out := env.List()

	Respond(res, req, http.StatusOK, out)
}

func GetHostHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	env, err := envs.Get(vars["env"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	host, err := env.Get(vars["host"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	Respond(res, req, http.StatusOK, host)
}

func ListServicesHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	env, err := envs.Get(vars["env"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	host, err := env.Get(vars["host"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	out := host.List()

	Respond(res, req, http.StatusOK, out)
}

func GetServiceHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	env, err := envs.Get(vars["env"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	host, err := env.Get(vars["host"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	service, err := host.Get(vars["service"])
	if err != nil {
		Respond(res, req, http.StatusNotFound, err.Error())
		return
	}

	Respond(res, req, http.StatusOK, service)
}

func UpdateServiceHandler(res http.ResponseWriter, req *http.Request) {
	var attrs = make(map[string]interface{})

	vars := mux.Vars(req)

	env := vars["env"]
	host := vars["host"]
	service := vars["service"]

	stateParam := req.URL.Query()["state"]
	if len(stateParam) > 0 {
		v, err := strconv.Atoi(stateParam[0])
		if err == nil {
			attrs["state"] = v
		}
	}

	ackParam := req.URL.Query()["acknowledgement"]
	if len(ackParam) > 0 {
		v, err := strconv.ParseBool(ackParam[0])
		if err == nil {
			attrs["acknowledgement"] = v
		}
	}

	downtimeParam := req.URL.Query()["downtime"]
	if len(downtimeParam) > 0 {
		v, err := strconv.ParseBool(downtimeParam[0])
		if err == nil {
			attrs["downtime"] = v
		}
	}

	instructions := Instruction{
		Env:     env,
		Host:    host,
		Service: service,
		Attrs:   attrs,
	}

	err := instructions.Apply()
	if err != nil {
		Respond(res, req, http.StatusInternalServerError, err.Error())
		return
	}

	message, err := json.Marshal(instructions)
	if err != nil {
		Respond(res, req, http.StatusInternalServerError, err.Error())
		return
	}

	hub.broadcast <- message
	Respond(res, req, http.StatusOK, service)
}
