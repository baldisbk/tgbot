package impl

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/baldisbk/tgbot_sample/pkg/statemachine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
)

const (
	listCallback         = "list"
	addCallback          = "add"
	reportCallback       = "report"
	postponeCallback     = "postpone"
	stopListCallback     = "stop_list"
	forwardListCallback  = "forward_list"
	backwardListCallback = "backward_list"
	displayCallback      = "display_"

	okCallback    = "ok"
	retryCallback = "retry"
	abortCallback = "abort"
)

func callbackResponse(input interface{}) (bool, string) {
	if input == nil {
		return false, ""
	}
	if rsp, ok := input.(*tgapi.CallbackQuery); ok {
		return true, rsp.Data
	}
	return false, ""
}

func checkCallback(check string) statemachine.SMPredicate {
	return func(ctx context.Context, state string, input interface{}) bool {
		ok, rsp := callbackResponse(input)
		return ok && (check == "*" || rsp == check)
	}
}

func (u *User) isStart(ctx context.Context, state string, input interface{}) bool {
	if input == nil {
		return false
	}
	rsp, ok := input.(*tgapi.Message)
	return ok && rsp.Text == "/start"
}

func (u *User) isTimer(ctx context.Context, state string, input interface{}) bool {
	if input == nil {
		return false
	}
	timer, ok := input.(*timer.TimerEvent)
	return ok && timer.Type == achievementTimer
}

func (u *User) isRollback(ctx context.Context, state string, input interface{}) bool {
	if input == nil {
		return false
	}
	switch event := input.(type) {
	case *timer.TimerEvent:
		return event.Type == timeoutTimer
	case *tgapi.Message:
		return event.Text == "/start"
	}
	return false
}

func (u *User) isDisplay(ctx context.Context, state string, input interface{}) bool {
	if input == nil {
		return false
	}
	if rsp, ok := input.(*tgapi.CallbackQuery); ok {
		ss := regexp.MustCompile(fmt.Sprintf("^%s([0-9]+)$", displayCallback)).FindStringSubmatch(rsp.Data)
		if len(ss) != 2 {
			return false
		}
		names, index := u.getNames()
		indPlus, err := strconv.Atoi(ss[1])
		if err != nil {
			return false
		}
		u.currentName = names[index+indPlus]
		return true
	}
	return false
}

func (u *User) isValidInput(ctx context.Context, state string, input interface{}) bool {
	if input == nil {
		return false
	}
	rsp, ok := input.(*tgapi.Message)
	switch state {
	case addState:
		switch u.stageNumber {
		case 0: //validate name
		case 1: //validate desc
		case 2: //validate limit
			if _, err := strconv.Atoi(rsp.Text); err != nil {
				return false
			}
		case 3: //validate current
			if _, err := strconv.Atoi(rsp.Text); err != nil {
				return false
			}
		}
	case reportState:
		if _, err := strconv.Atoi(rsp.Text); err != nil {
			return false
		}
	}
	return ok
}
