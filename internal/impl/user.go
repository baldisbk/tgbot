package impl

import (
	"context"
	"time"

	"github.com/baldisbk/tgbot/pkg/statemachine"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/baldisbk/tgbot/pkg/timer"
)

const (
	achievementTimer = "achievement"
	timeoutTimer     = "timeout"
)

type User struct {
	Id      uint64
	Name    string
	Limits  map[string]*LimitAchievement
	Strikes map[string]*StrikeAchievement

	// settings
	dialogTimeout time.Duration

	// internals
	tgClient tgapi.TGClient
	timer    *timer.Timer
	machine  statemachine.Machine

	// dialog state
	currentName string
	lastMessage uint64
	stageNumber int
	newLimit    *LimitAchievement // add limit
}

// probably nothing needed
func (u *User) UpdateState(context.Context, interface{}) error { return nil }
func (u *User) Run(ctx context.Context, input interface{}) (interface{}, error) {
	return u.machine.Run(ctx, input)
}

func (u *User) SetTimer(name string, t time.Time) {
	u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, name, achievementTimer, t)
}
func (u *User) SetTimeout() {
	u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name},
		"timeout", achievementTimer, time.Now().Add(u.dialogTimeout))
}

func (u *User) Wake() {
	for name, limit := range u.Limits {
		u.SetTimer(name, limit.CheckTime)
	}
	for name, strike := range u.Strikes {
		u.SetTimer(name, strike.CheckTime)
	}
}
