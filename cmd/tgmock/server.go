package main

import (
	"io/ioutil"
	"net/http"

	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/gorilla/mux"
)

type Server struct {
	messages []tgapi.Update
}

func (s *Server) dflt(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.S(r.Context()).Errorf("serve err: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	logging.S(r.Context()).Infof("served dflt %s %s %s", r.URL, string(cts), vars["token"])
	rw.Write([]byte("{}"))
	return
}

func (s *Server) update(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.S(r.Context()).Errorf("serve err: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	logging.S(r.Context()).Infof("served get-update %s", string(cts))
	rw.Write([]byte("{}"))
	return
}

func (s *Server) message(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.S(r.Context()).Errorf("serve err: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	logging.S(r.Context()).Infof("served msg %s", string(cts))
	rw.Write([]byte("{}"))
	return
}

func (s *Server) callback(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.S(r.Context()).Errorf("serve err: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	logging.S(r.Context()).Infof("served callback %s", string(cts))
	rw.Write([]byte("{}"))
	return
}
