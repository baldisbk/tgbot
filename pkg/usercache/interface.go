package usercache

import (
	"context"

	"github.com/baldisbk/tgbot/pkg/tgapi"
)

type User interface {
	UpdateState(context.Context, interface{}) error
	Run(ctx context.Context, input interface{}) (interface{}, error)
}

type UserCache interface {
	Get(context.Context, tgapi.User) (User, error)
	Put(context.Context, tgapi.User, User) error
	Close()
}

type UserFactory interface {
	MakeUser(context.Context, tgapi.User) User
}
