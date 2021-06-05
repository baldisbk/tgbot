package engine

import (
	"github.com/baldisbk/tgbot_sample/internal/tgapi"
	"github.com/baldisbk/tgbot_sample/internal/usercache"

	"golang.org/x/xerrors"
)

type Engine struct {
	client *tgapi.TGClient
	users  map[uint64]usercache.User
	cache  usercache.UserCache
}

func NewEngine(client *tgapi.TGClient, cache usercache.UserCache) *Engine {
	return &Engine{
		users:  map[uint64]usercache.User{},
		cache:  cache,
		client: client,
	}
}

func (e *Engine) Receive(update tgapi.Update) error {
	var tgUser tgapi.User
	var user usercache.User
	var ok bool
	var err error
	switch {
	case update.Message != nil:
		tgUser = update.Message.From
	case update.CallbackQuery != nil:
		tgUser = update.CallbackQuery.From
	}
	if user, ok = e.users[tgUser.Id]; !ok {
		var err error
		if user, err = e.cache.Get(tgUser); err != nil {
			// database problem
			return xerrors.Errorf("get user from cache: %w", err)
		}
	}
	var rsp interface{}
	switch {
	case update.Message != nil:
		rsp, err = user.Machine().Run(update.Message)
		if err != nil {
			// retriable (network)
			return xerrors.Errorf("process message: %w", err)
		}
	case update.CallbackQuery != nil:
		_, err := user.Machine().Run(update.CallbackQuery)
		if err != nil {
			// retriable (network)
			return xerrors.Errorf("process callback: %w", err)
		}
		if err := e.client.AnswerCallback(update.CallbackQuery.Id); err != nil {
			// retriable (network)
			return xerrors.Errorf("confirm callback: %w", err)
		}
	}

	if err := user.UpdateState(rsp); err != nil {
		// bad response
		return xerrors.Errorf("update user state: %w", err)
	}
	if err := e.cache.Put(tgUser, user); err != nil {
		// database problem
		return xerrors.Errorf("put user to cache: %w", err)
	}

	return nil
}
