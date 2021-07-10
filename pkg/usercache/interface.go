package usercache

import (
	"context"

	"github.com/baldisbk/tgbot_sample/pkg/statemachine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

type User interface {
	UpdateState(context.Context,interface{}) error
	Machine() statemachine.Machine
}

type UserCache interface {
	Get(context.Context, tgapi.User) (User, error)
	Put(context.Context, tgapi.User, User) error
	Close()
}

type UserFactory interface {
	MakeUser(context.Context, tgapi.User) User
}
