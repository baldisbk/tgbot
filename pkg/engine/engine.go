package engine

import (
	"context"

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
	PreProcess(ctx context.Context, client *tgapi.TGClient) error
	PostProcess(ctx context.Context, client *tgapi.TGClient) error
}

func NewEngine(client *tgapi.TGClient, cache usercache.UserCache) *Engine {
	return &Engine{
		users:  map[uint64]usercache.User{},
		cache:  cache,
		client: client,
	}
}

func (e *Engine) Receive(ctx context.Context, signal Signal) error {
	tgUser := signal.User()
	user, ok := e.users[tgUser.Id]
	var err error
	if !ok {
		if user, err = e.cache.Get(ctx, tgUser); err != nil {
			// database problem
			return xerrors.Errorf("get user from cache: %w", err)
		}
	}
	if err := signal.PreProcess(ctx, e.client); err != nil {
		// retriable (network)
		return xerrors.Errorf("preprocess signal: %w", err)
	}
	rsp, err := user.Machine().Run(ctx, signal.Message())
	if err != nil {
		// retriable (network)
		return xerrors.Errorf("process signal: %w", err)
	}
	if err := signal.PostProcess(ctx, e.client); err != nil {
		// retriable (network)
		return xerrors.Errorf("postprocess signal: %w", err)
	}

	if err := user.UpdateState(ctx, rsp); err != nil {
		// bad response
		return xerrors.Errorf("update user state: %w", err)
	}
	if err := e.cache.Put(ctx, tgUser, user); err != nil {
		// database problem
		return xerrors.Errorf("put user to cache: %w", err)
	}

	return nil
}
