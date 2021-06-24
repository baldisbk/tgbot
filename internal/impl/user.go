package impl

import (
	"github.com/baldisbk/tgbot_sample/pkg/statemachine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
)

type User struct {
	Id      uint64
	Name    string
	Limits  map[string]*LimitAchievement
	Strikes map[string]*StrikeAchievement

	// internals
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
func (u *User) UpdateState(interface{}) error { return nil }
func (u *User) Machine() statemachine.Machine { return u.machine }
func (u *User) Wake() {
	for name, limit := range u.Limits {
		u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, name, limit.CheckTime)
	}
	for name, strike := range u.Strikes {
		u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, name, strike.CheckTime)
	}
}
