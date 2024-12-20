package engine

import (
	"context"
	"strconv"

	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/baldisbk/tgbot/pkg/usercache"

	"golang.org/x/xerrors"
)

type Engine interface {
	Receive(ctx context.Context, signal Signal) error
}

type Signal interface {
	User() tgapi.User
	Message() interface{}
	PreProcess(ctx context.Context, client tgapi.TGClient) error
	PostProcess(ctx context.Context, client tgapi.TGClient) error
}

type engine struct {
	client tgapi.TGClient
	cache  usercache.UserCache
}

func NewEngine(client tgapi.TGClient, cache usercache.UserCache) *engine {
	return &engine{
		cache:  cache,
		client: client,
	}
}

func (e *engine) Receive(ctx context.Context, signal Signal) error {
	tgUser := signal.User()
	ctx = logging.WithTag(ctx, "USER", strconv.FormatUint(tgUser.Id, 16))
	var err error
	var user usercache.User
	if user, err = e.cache.Get(ctx, tgUser); err != nil {
		// database problem
		return xerrors.Errorf("get user from cache: %w", err)
	}
	logging.S(ctx).Infof("Received signal (%#v) for user %v", signal, user)
	if err := signal.PreProcess(ctx, e.client); err != nil {
		// retriable (network)
		return xerrors.Errorf("preprocess signal: %w", err)
	}
	rsp, err := user.Run(ctx, signal.Message())
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
