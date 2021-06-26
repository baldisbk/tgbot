package engine

import (
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/usercache"

	"golang.org/x/xerrors"
)

type Engine struct {
	client *tgapi.TGClient
	users  map[uint64]usercache.User
	cache  usercache.UserCache
}

type Signal interface {
	User() tgapi.User
	Message() interface{}
	PreProcess(client *tgapi.TGClient) error
	PostProcess(client *tgapi.TGClient) error
}

func NewEngine(client *tgapi.TGClient, cache usercache.UserCache) *Engine {
	return &Engine{
		users:  map[uint64]usercache.User{},
		cache:  cache,
		client: client,
	}
}

func (e *Engine) Receive(signal Signal) error {
	tgUser := signal.User()
	user, ok := e.users[tgUser.Id]
	var err error
	if !ok {
		if user, err = e.cache.Get(tgUser); err != nil {
			// database problem
			return xerrors.Errorf("get user from cache: %w", err)
		}
	}
	if err := signal.PreProcess(e.client); err != nil {
		// retriable (network)
		return xerrors.Errorf("preprocess signal: %w", err)
	}
	rsp, err := user.Machine().Run(signal.Message())
	if err != nil {
		// retriable (network)
		return xerrors.Errorf("process signal: %w", err)
	}
	if err := signal.PostProcess(e.client); err != nil {
		// retriable (network)
		return xerrors.Errorf("postprocess signal: %w", err)
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
