package impl

import (
	"github.com/baldisbk/tgbot_sample/internal/usercache"
	"github.com/baldisbk/tgbot_sample/pkg/statemachine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
)

type user struct {
	Id      uint64
	Name    string
	Limits  map[string]*LimitAchievement
	Strikes map[string]*StrikeAchievement

	tgClient *tgapi.TGClient
	timer    *timer.Timer
	machine  statemachine.Machine

	// dialog state
	currentName string
	lastMessage uint64
	stageNumber int

	// add limit
	newLimit *LimitAchievement
}

// probably nothing needed
func (u *user) UpdateState(interface{}) error { return nil }
func (u *user) Machine() statemachine.Machine { return u.machine }
func (u *user) Wake() {
	for name, limit := range u.Limits {
		u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, name, limit.CheckTime)
	}
	for name, strike := range u.Strikes {
		u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, name, strike.CheckTime)
	}
}

type userFactory struct {
	tgClient *tgapi.TGClient
	timer    *timer.Timer
}

func NewFactory(tgClient *tgapi.TGClient, timer *timer.Timer) *userFactory {
	return &userFactory{tgClient: tgClient, timer: timer}
}

func (f *userFactory) MakeUser(u tgapi.User) usercache.User {
	res := &user{
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
