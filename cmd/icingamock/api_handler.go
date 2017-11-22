package main

import (
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

	stateParam := req.URL.Query()["state"]
	if len(stateParam) > 0 {
		i, err := strconv.Atoi(stateParam[0])
		if err == nil {
			service.CheckState = i
		}
	}

	Respond(res, req, http.StatusOK, service)
}
