package tgmock

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/baldisbk/tgbot/internal/config"
	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/gorilla/mux"
)

const (
	privateMessagePath = "/private/message"
	privateButtonPath  = "/private/button"
	privateHistoryPath = "/private/history"
)

type HistoryEntry struct {
	UserID   uint64 `json:"user_id"`
	FromUser string
}

type Server struct {
	messages []tgapi.Update
}

type Config struct {
	config.ConfigFlags

	Address string `yaml:"address"`
}

func NewServer(ctx context.Context, cfg Config) *http.Server {
	srv := Server{}

	mx := mux.NewRouter()
	mx.HandleFunc("/{token}/"+tgapi.TestCmd, srv.ping)
	mx.HandleFunc("/{token}/"+tgapi.ReceiveCmd, srv.update)
	mx.HandleFunc("/{token}/"+tgapi.SendCmd, srv.message)
	mx.HandleFunc("/{token}/"+tgapi.AnswerCmd, srv.callback)
	mx.HandleFunc("/{token}/"+tgapi.EditCmd, srv.message)

	mx.HandleFunc(privateMessagePath, srv.privateMessage)
	mx.HandleFunc(privateButtonPath, srv.privateButton)
	mx.HandleFunc(privateHistoryPath, srv.privateHistory)

	mx.NotFoundHandler = http.HandlerFunc(srv.dflt)

	return &http.Server{
		Addr:        cfg.Address,
		Handler:     mx,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
}

type errorMsg struct {
	Message string `json:"message"`
}

func (s *Server) writeError(rw http.ResponseWriter, r *http.Request, code int, msg string, err error) {
	output := fmt.Sprintf(msg, err)
	logging.S(r.Context()).Errorf(output)

	b, e := json.Marshal(errorMsg{Message: output})
	if e != nil {
		logging.S(r.Context()).Errorf("marshal: %s", e)
		return
	}
	rw.WriteHeader(code)
	_, err = rw.Write(b)
	if err != nil {
		logging.S(r.Context()).Errorf("write: %s", err)
	}
}

func (s *Server) dflt(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "serve err: %s", err)
		return
	}
	vars := mux.Vars(r)
	logging.S(r.Context()).Infof("??? %s %s %s", r.URL, string(cts), vars["token"])
	rw.Write([]byte("{}"))
	return
}

func (s *Server) ping(rw http.ResponseWriter, r *http.Request) {
	_, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "serve err: %s", err)
		return
	}
	vars := mux.Vars(r)
	logging.S(r.Context()).Infof("ping %s", vars["token"])
	rw.Write([]byte("{}"))
	return
}

func (s *Server) update(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "serve err: %s", err)
		return
	}
	var payload tgapi.GetUpdates
	err = json.Unmarshal(cts, &payload)
	if err != nil {
		s.writeError(rw, r, http.StatusBadRequest, "bad request: %s", err)
		return
	}
	logging.S(r.Context()).Infof("--- get-update %s", string(cts))
	var messages []tgapi.Update
	for _, up := range s.messages {
		if up.UpdateId >= payload.Offset {
			messages = append(messages, up)
		}
	}
	b, err := json.Marshal(tgapi.UpdateResponse{Result: messages, Ok: true})
	if err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "marshal err: %s", err)
		return
	}
	if _, err := rw.Write(b); err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "write err: %s", err)
		return
	}
}

func (s *Server) message(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	logging.S(r.Context()).Infof("< bot < : %s", string(cts))
	rw.Write([]byte("{}"))
	return
}

func (s *Server) callback(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	logging.S(r.Context()).Infof("<- : %s", string(cts))
	rw.Write([]byte("{}"))
	return
}
