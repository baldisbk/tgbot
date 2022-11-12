package tgmock

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

type PrivateRequest struct {
	UserID  uint64 `json:"user_id"`
	Payload string `json:"payload,omitempty"`
}

func (s *Server) privateMessage(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "serve err: %s", err)
		return
	}
	var payload PrivateRequest
	err = json.Unmarshal(cts, &payload)
	if err != nil {
		s.writeError(rw, r, http.StatusBadRequest, "bad request: %s", err)
		return
	}
	logging.S(r.Context()).Infof("> usr %d > : %s", payload.UserID, payload.Payload)
	var id uint64
	if len(s.messages) != 0 {
		id = s.messages[len(s.messages)-1].UpdateId + 1
	}
	s.messages = append(s.messages, tgapi.Update{
		UpdateId: id,
		Message: &tgapi.Message{
			MessageId: id,
			From:      tgapi.User{Id: payload.UserID, FirstName: "Test user"},
			Chat:      tgapi.Chat{},
			Text:      payload.Payload,
		},
	})
	rw.WriteHeader(http.StatusOK)
	return
}

func (s *Server) privateButton(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "serve err: %s", err)
		return
	}
	var payload PrivateRequest
	err = json.Unmarshal(cts, &payload)
	if err != nil {
		s.writeError(rw, r, http.StatusBadRequest, "bad request: %s", err)
		return
	}
	logging.S(r.Context()).Infof("> usr %d > + %s", payload.UserID, payload.Payload)
	var id uint64
	if len(s.messages) != 0 {
		id = s.messages[len(s.messages)-1].UpdateId + 1
	}
	s.messages = append(s.messages, tgapi.Update{
		UpdateId: id,
		CallbackQuery: &tgapi.CallbackQuery{
			From: tgapi.User{Id: payload.UserID, FirstName: "Test user"},
			Data: payload.Payload,
		},
	})
	rw.WriteHeader(http.StatusOK)
	return
}

func (s *Server) privateHistory(rw http.ResponseWriter, r *http.Request) {
	cts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeError(rw, r, http.StatusInternalServerError, "serve err: %s", err)
		return
	}
	var payload PrivateRequest
	err = json.Unmarshal(cts, &payload)
	if err != nil {
		s.writeError(rw, r, http.StatusBadRequest, "bad request: %s", err)
		return
	}
	logging.S(r.Context()).Infof("==== HISTORY %d", payload.UserID)
	var id uint64
	if len(s.messages) != 0 {
		id = s.messages[len(s.messages)-1].UpdateId + 1
	}
	s.messages = append(s.messages, tgapi.Update{
		UpdateId: id,
		CallbackQuery: &tgapi.CallbackQuery{
			From: tgapi.User{FirstName: "Test user"},
			Data: string(cts),
		},
	})
	rw.WriteHeader(http.StatusOK)
	return
}
