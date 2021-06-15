package impl

import (
	"github.com/baldisbk/tgbot_sample/internal/statemachine"
	"github.com/baldisbk/tgbot_sample/internal/tgapi"
	"github.com/baldisbk/tgbot_sample/internal/usercache"
)

type user struct {
	Id      uint64
	Name    string
	Limits  map[string]*LimitAchievement
	Strikes map[string]*StrikeAchievement

	tgClient *tgapi.TGClient
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

type userFactory struct {
	tgClient *tgapi.TGClient
}

func NewFactory(tgClient *tgapi.TGClient) *userFactory {
	return &userFactory{tgClient: tgClient}
}

func (f *userFactory) Factory(u tgapi.User) usercache.User {
	res := &user{
		Id:   u.Id,
		Name: u.FirstName,

		Limits:  map[string]*LimitAchievement{},
		Strikes: map[string]*StrikeAchievement{},

		tgClient: f.tgClient,
	}
	res.machine = statemachine.NewSM(startState, makeTransitions(res))
	return res
}
