package impl

import (
	"time"

	"github.com/baldisbk/tgbot_sample/pkg/statemachine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
)

type userFactory struct {
	tgClient tgapi.TGClient
	timer    *timer.Timer

	config Config
}

type Config struct {
	DialogTimeout time.Duration `yaml:"dialog_timeout"`
}

func NewFactory(cfg Config, tgClient tgapi.TGClient, timer *timer.Timer) *userFactory {
	return &userFactory{config: cfg, tgClient: tgClient, timer: timer}
}

func (f *userFactory) MakeUser(u tgapi.User) *User {
	res := &User{
		Id:   u.Id,
		Name: u.FirstName,

		Limits:  map[string]*LimitAchievement{},
		Strikes: map[string]*StrikeAchievement{},

		tgClient: f.tgClient,
		timer:    f.timer,

		dialogTimeout: f.config.DialogTimeout,
	}
	res.machine = statemachine.NewSM(startState, makeTransitions(res))
	return res
}
