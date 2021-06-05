package impl

import (
	"fmt"

	"github.com/baldisbk/tgbot_sample/internal/statemachine"
	"github.com/baldisbk/tgbot_sample/internal/tgapi"
	"github.com/baldisbk/tgbot_sample/internal/usercache"
	"golang.org/x/xerrors"
)

type user struct {
	Id      uint64
	Name    string
	Counter int

	tgClient *tgapi.TGClient
	machine  statemachine.Machine
}

// type SMPredicate func(string, interface{}) bool

func (u *user) isMessage(state string, input interface{}) bool {
	if input == nil {
		return false
	}
	rsp, ok := input.(*tgapi.Message)
	return ok && rsp.Text != "MAGIC"
}

func (u *user) isMagic(state string, input interface{}) bool {
	if input == nil {
		return false
	}
	rsp, ok := input.(*tgapi.Message)
	return ok && rsp.Text == "MAGIC"
}

func (u *user) isCallback(state string, input interface{}) bool {
	if input == nil {
		return false
	}
	_, ok := input.(*tgapi.CallbackQuery)
	return ok && state == "keyboard"
}

// type SMCallback func(interface{}) (interface{}, error)

func (u *user) doMessage(input interface{}) (interface{}, error) {
	rsp := input.(*tgapi.Message)
	u.Counter++
	if err := u.tgClient.SendMessage(u.Id, fmt.Sprintf("We've got \"%s\"", rsp.Text)); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	return nil, nil
}

func (u *user) doKeyboard(input interface{}) (interface{}, error) {
	if err := u.tgClient.SendMessage(u.Id, fmt.Sprintf("We've got MAGIC WORD")); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	if err := u.tgClient.CreateInputKeyboard(u.Id,
		fmt.Sprintf("MAGIC for %s", u.Name),
		tgapi.InlineKeyboard{
			InlineKeyboard: [][]tgapi.InlineKeyboardButton{
				{{Text: "first", CallbackData: "1"}, {Text: "second", CallbackData: "2"}},
				{{Text: "third", CallbackData: "3"}, {Text: "fourth", CallbackData: "4"}},
			},
		}); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	return nil, nil
}

func (u *user) doCallback(input interface{}) (interface{}, error) {
	rsp, ok := input.(*tgapi.CallbackQuery)
	if !ok {
		return nil, xerrors.Errorf("%T is not a callback", input)
	}
	if err := u.tgClient.DropKeyboard(u.Id, fmt.Sprintf("Callback %s", rsp.Data)); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	if err := u.tgClient.SendMessage(u.Id, fmt.Sprintf("Counter %d", u.Counter)); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	return nil, nil
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
		Id:       u.Id,
		Name:     u.FirstName,
		tgClient: f.tgClient,
	}
	res.machine = statemachine.NewSM("start", []statemachine.Transition{
		{
			Source: "start", Destination: "start",
			Predicate: res.isMessage,
			Callback:  res.doMessage,
		},
		{
			Source: "start", Destination: "keyboard",
			Predicate: res.isMagic,
			Callback:  res.doKeyboard,
		},
		{
			Source: "keyboard", Destination: "start",
			Predicate: res.isCallback,
			Callback:  res.doCallback,
		},
	})
	return res
}
