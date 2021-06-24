package impl

import (
	"github.com/baldisbk/tgbot_sample/pkg/statemachine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
)

type userFactory struct {
	tgClient *tgapi.TGClient
	timer    *timer.Timer
}

func NewFactory(tgClient *tgapi.TGClient, timer *timer.Timer) *userFactory {
	return &userFactory{tgClient: tgClient, timer: timer}
}

func (f *userFactory) MakeUser(u tgapi.User) *User {
	res := &User{
		Id:   u.Id,
		Name: u.FirstName,

		Limits:  map[string]*LimitAchievement{},
		Strikes: map[string]*StrikeAchievement{},

		tgClient: f.tgClient,
		timer:    f.timer,
	}
	res.machine = statemachine.NewSM(startState, makeTransitions(res))
	return res
}
