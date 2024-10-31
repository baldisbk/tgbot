package poller

import (
	"context"
	"testing"
	"time"

	"github.com/baldisbk/tgbot/pkg/engine"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/mock"
)

func TestPoller(t *testing.T) {
	testCases := []struct {
		desc   string
		update tgapi.Update
	}{
		{
			desc: "message",
			update: tgapi.Update{
				Message: &tgapi.Message{Text: "text"},
			},
		},
		{
			desc: "call",
			update: tgapi.Update{
				CallbackQuery: &tgapi.CallbackQuery{Data: "data"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			clock := clockwork.NewFakeClock()
			engine := engine.NewEngineMock()

			updates := []tgapi.Update{{}}
			tgClient := tgapi.NewMock()

			poller := newPoller(ctx, Config{PollPeriod: time.Second}, clock, tgClient, engine)
			defer poller.Shutdown()
			// nothing called at all

			tgClient.On(
				"GetUpdates",
				mock.Anything,
			).Return(updates, nil).Twice()

			clock.Advance(time.Second)
			time.Sleep(time.Millisecond)
			// client called, engine is not

			updates[0] = tC.update
			if tC.update.Message != nil {
				engine.On(
					"Receive",
					mock.Anything,
					mock.MatchedBy(func(e *tgapi.Message) bool { return e.Text == "text" }),
				).Return(nil).Once()
			}
			if tC.update.CallbackQuery != nil {
				engine.On(
					"Receive",
					mock.Anything,
					mock.MatchedBy(func(e *tgapi.CallbackQuery) bool { return e.Data == "data" }),
				).Return(nil).Once()
			}
			clock.Advance(time.Second)
			time.Sleep(time.Millisecond)
			// engine called
			engine.AssertExpectations(t)
		})
	}
}
