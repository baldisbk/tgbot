package usercache

import (
	"github.com/baldisbk/tgbot_sample/internal/statemachine"
	"github.com/baldisbk/tgbot_sample/internal/tgapi"
)

type User interface {
	UpdateState(interface{}) error
	Machine() statemachine.Machine
}

type UserCache interface {
	Get(user tgapi.User) (User, error)
	Put(user tgapi.User, state User) error
	Close()
}

type UserFactory interface {
	MakeUser(tgapi.User) User
}
