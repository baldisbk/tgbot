package impl

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/baldisbk/tgbot_sample/internal/tgapi"
	"github.com/baldisbk/tgbot_sample/internal/timer"
	"golang.org/x/xerrors"
)

const listLength = 5

func (u *user) ask(message string, options []tgapi.InlineKeyboardButton) (interface{}, error) {
	if msgId, err := u.tgClient.EditInputKeyboard(u.Id, message, u.lastMessage,
		tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{options}}); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	} else {
		u.lastMessage = msgId
	}
	return nil, nil
}

func (u *user) doNoUnderstand(input interface{}) (interface{}, error) {
	message := fmt.Sprintf("Can't understand you")
	if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	u.lastMessage = 0
	return nil, nil
}

func (u *user) doStart(input interface{}) (interface{}, error) {
	// drop state to defaults
	names, _ := u.getNames()
	if len(names) != 0 {
		u.currentName = names[0]
	} else {
		u.currentName = ""
	}
	u.stageNumber = 0
	// menu
	message := fmt.Sprintf("Hello, %s, whacha gonna do?", u.Name)
	return u.ask(message, []tgapi.InlineKeyboardButton{
		{Text: "Display achievements", CallbackData: listCallback},
		{Text: "Add achievement", CallbackData: addCallback},
	})
}

func (u *user) doTimer(input interface{}) (interface{}, error) {
	rsp := input.(*timer.TimerEvent)
	message := fmt.Sprintf("Time has come to report progress of %s", rsp.Name)
	return u.ask(message, []tgapi.InlineKeyboardButton{
		{Text: "Let's go", CallbackData: reportCallback},
		{Text: "Later...", CallbackData: postponeCallback},
	})
}

func (u *user) getNames() ([]string, int) {
	names := []string{}
	for name := range u.Limits {
		names = append(names, name)
	}
	for name := range u.Strikes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, sort.SearchStrings(names, u.currentName)
}

func (u *user) doList(input interface{}) (interface{}, error) {
	names, index := u.getNames()
	if len(names) == 0 {
		message := fmt.Sprintf("Nothing to display")
		if msgId, err := u.tgClient.EditInputKeyboard(u.Id, message, u.lastMessage,
			tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
				{{Text: "Back", CallbackData: stopListCallback}},
			}}); err != nil {
			return nil, xerrors.Errorf("send: %w", err)
		} else {
			u.lastMessage = msgId
		}
		return nil, nil
	}
	message := fmt.Sprintf("What to display")
	keyboard := [][]tgapi.InlineKeyboardButton{}
	for i := 0; i < listLength && i+index < len(names); i++ {
		// TODO show progress as well
		keyboard = append(keyboard, []tgapi.InlineKeyboardButton{{
			Text:         names[i+index],
			CallbackData: fmt.Sprintf(displayCallback+"%d", i),
		}})
	}
	controls := []tgapi.InlineKeyboardButton{}
	if index > 0 {
		controls = append(controls, tgapi.InlineKeyboardButton{Text: "<", CallbackData: backwardListCallback})
	}
	controls = append(controls, tgapi.InlineKeyboardButton{Text: "Back", CallbackData: stopListCallback})
	if index+listLength < len(names) {
		controls = append(controls, tgapi.InlineKeyboardButton{Text: ">", CallbackData: forwardListCallback})
	}
	if msgId, err := u.tgClient.EditInputKeyboard(u.Id, message, u.lastMessage,
		tgapi.InlineKeyboard{InlineKeyboard: append(keyboard, controls)},
	); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	} else {
		u.lastMessage = msgId
	}
	return nil, nil
}

func (u *user) doListForward(input interface{}) (interface{}, error) {
	names, index := u.getNames()
	if index+listLength < len(names) {
		index += listLength
	}
	u.currentName = names[index]
	return u.doList(input)
}

func (u *user) doListBackward(input interface{}) (interface{}, error) {
	names, index := u.getNames()
	if index-listLength < 0 {
		index = 0
	} else {
		index -= listLength
	}
	u.currentName = names[index]
	return u.doList(input)
}

func (u *user) doDisplay(input interface{}) (interface{}, error) {
	var message string
	if limit, ok := u.Limits[u.currentName]; ok {
		if limit.Done {
			message = fmt.Sprintf("%s.\nAchivement DONE!\n%s", limit.Name, limit.Description)
		} else {
			var required, achieved int
			if limit.Ascend {
				required = limit.Limit - limit.Initial
				achieved = limit.Current - limit.Initial
			} else {
				required = limit.Initial - limit.Limit
				achieved = limit.Initial - limit.Current
			}
			message = fmt.Sprintf("%s.\nAchivement progress: %.2f%% (%d/%d)\n%s",
				limit.Name, (float32(achieved)/float32(required))*100, limit.Current, limit.Initial, limit.Description)
		}
	} else if strike, ok := u.Strikes[u.currentName]; ok {
		if strike.Done {
			message = fmt.Sprintf("%s.\nAchivement DONE!\n%s", strike.Name, strike.Description)
		} else {
			message = fmt.Sprintf("%s.\nAchivement progress: %.2f%% (%d/%d, best %d)\n%s",
				strike.Name, (float32(strike.Last)/float32(strike.Strike))*100,
				strike.Last, strike.Strike, strike.Best, strike.Description)
		}
	} else {
		return nil, xerrors.Errorf("unexpected achivement name: %s", u.currentName)
	}
	return u.ask(message, []tgapi.InlineKeyboardButton{
		{Text: "Back to list", CallbackData: listCallback},
		{Text: "Back to menu", CallbackData: stopListCallback},
	})
}

func (u *user) doPostpone(input interface{}) (interface{}, error) {
	// TODO custom postpone time via menu
	// now postpone to 3 hour
	if limit, ok := u.Limits[u.currentName]; ok {
		limit.CheckTime = time.Now().Add(3 * time.Hour)
	} else if strike, ok := u.Strikes[u.currentName]; ok {
		strike.CheckTime = time.Now().Add(3 * time.Hour)
	} else {
		return nil, xerrors.Errorf("unexpected achivement name: %s", u.currentName)
	}
	return nil, nil
}

func (u *user) doStartAdd(input interface{}) (interface{}, error) {
	message := fmt.Sprintf("Okay, now would you enter achievement name")
	if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	u.stageNumber = 0
	u.newLimit = &LimitAchievement{}
	return nil, nil
}

func (u *user) doAdd(input interface{}) (interface{}, error) {
	rsp := input.(*tgapi.Message)
	switch u.stageNumber {
	case 0:
		u.newLimit.Name = rsp.Text
		message := fmt.Sprintf("Now would you enter achievement description")
		if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
			return nil, xerrors.Errorf("send: %w", err)
		}
	case 1:
		u.newLimit.Description = rsp.Text
		message := fmt.Sprintf("Now what about limit to achieve?")
		if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
			return nil, xerrors.Errorf("send: %w", err)
		}
	case 2:
		val, _ := strconv.Atoi(rsp.Text)
		u.newLimit.Limit = val
		message := fmt.Sprintf("Okay, and where are you now?")
		if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
			return nil, xerrors.Errorf("send: %w", err)
		}
	case 3:
		val, _ := strconv.Atoi(rsp.Text)
		u.newLimit.Initial = val
		u.lastMessage = 0
		u.stageNumber = 0
		message := fmt.Sprintf("So, you are to add %s, OK?", u.newLimit.Name)
		return u.ask(message, []tgapi.InlineKeyboardButton{
			{Text: "OK", CallbackData: okCallback},
			{Text: "Fix it", CallbackData: retryCallback},
			{Text: "Fuck it", CallbackData: abortCallback},
		})
	}
	u.stageNumber++
	return nil, nil
}

func (u *user) doFinishAdd(input interface{}) (interface{}, error) {
	u.newLimit.Ascend = u.newLimit.Limit > u.newLimit.Initial
	u.newLimit.Current = u.newLimit.Initial
	u.newLimit.CheckTime = time.Now().Add(24 * time.Hour)
	u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, u.newLimit.Name, u.newLimit.CheckTime)
	u.Limits[u.newLimit.Name] = u.newLimit
	u.newLimit = nil
	u.lastMessage = 0
	return nil, nil
}

func (u *user) doReport(input interface{}) (interface{}, error) {
	timer := input.(*timer.TimerEvent)
	message := fmt.Sprintf("Okay, now would you enter current state of %s", timer.Name)
	if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	u.currentName = timer.Name
	return nil, nil
}

func (u *user) doFinishReport(input interface{}) (interface{}, error) {
	rsp := input.(*tgapi.Message)
	val, _ := strconv.Atoi(rsp.Text)
	if limit, ok := u.Limits[u.currentName]; ok {
		limit.Current = val
		if limit.Ascend && limit.Current >= limit.Limit {
			limit.Done = true
		}
		if !limit.Ascend && limit.Current <= limit.Limit {
			limit.Done = true
		}
		if limit.Done {
			message := fmt.Sprintf("Wow, you've done it! Gratz! Achievement %s completed!", u.currentName)
			if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
				return nil, xerrors.Errorf("send: %w", err)
			}
		}
		limit.CheckTime = limit.CheckTime.Add(24 * time.Hour)
		u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, limit.Name, limit.CheckTime)
	}
	if strike, ok := u.Strikes[u.currentName]; ok {
		if strike.Ascend && val >= strike.Limit {
			strike.Last++
		} else if !strike.Ascend && val <= strike.Limit {
			strike.Last++
		} else {
			if strike.Last > strike.Best {
				strike.Best = strike.Last
			}
			strike.Last = 0
		}
		if strike.Last >= strike.Strike {
			strike.Done = true
			message := fmt.Sprintf("Wow, you've done it! Gratz! Achievement %s completed!", u.currentName)
			if _, err := u.tgClient.SendMessage(u.Id, message); err != nil {
				return nil, xerrors.Errorf("send: %w", err)
			}
		}
		strike.CheckTime = strike.CheckTime.Add(24 * time.Hour)
		u.timer.SetAlarm(tgapi.User{Id: u.Id, FirstName: u.Name}, strike.Name, strike.CheckTime)
	}
	return nil, nil
}
