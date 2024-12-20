package timer

import (
	"context"
	"testing"
	"time"

	"github.com/baldisbk/tgbot/pkg/engine"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTimer(t *testing.T) {
	testCases := []struct {
		desc           string
		duration       time.Duration
		alarm1, alarm2 time.Duration
		exp1, exp2     bool
	}{
		{
			desc:     "single",
			alarm1:   1 * time.Second,
			duration: 1 * time.Second,
			exp1:     true,
		},
		{
			desc:     "one",
			alarm1:   1 * time.Second,
			alarm2:   2 * time.Second,
			duration: 1 * time.Second,
			exp1:     true,
		},
		{
			desc:     "both",
			alarm1:   1 * time.Second,
			alarm2:   1 * time.Second,
			duration: 1 * time.Second,
			exp1:     true,
			exp2:     true,
		},
	}
	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			assert := require.New(t)
			ctx := context.Background()

			clock := clockwork.NewFakeClock()
			alarm1 := clock.Now().Add(c.alarm1)
			alarm2 := clock.Now().Add(c.alarm2)
			duration := clock.Now().Add(c.duration)

			user1 := tgapi.User{Id: 1}
			user2 := tgapi.User{Id: 2}

			var received1, received2 bool
			engine := engine.NewEngineMock()
			engine.On(
				"Receive",
				mock.Anything,
				mock.MatchedBy(func(e *TimerEvent) bool {
					return e.Type == "1" &&
						e.Name == "1" &&
						e.Receiver == user1 &&
						e.Time == alarm1
				})).
				Return(nil).
				Run(func(args mock.Arguments) { received1 = true })
			engine.On(
				"Receive",
				mock.Anything,
				mock.MatchedBy(func(e *TimerEvent) bool {
					return e.Type == "2" &&
						e.Name == "2" &&
						e.Receiver == user2 &&
						e.Time == alarm2
				})).
				Return(nil).
				Run(func(args mock.Arguments) { received2 = true })

			timer := newTimer(ctx, engine, clock, time.Second)
			defer timer.Shutdown()

			if c.alarm1 != 0 {
				timer.SetAlarm(user1, "1", "1", alarm1)
			}
			if c.alarm2 != 0 {
				timer.SetAlarm(user2, "2", "2", alarm2)
			}

			for clock.Now().Before(duration) {
				assert.False(received1)
				assert.False(received2)
				timer.advance(time.Second)
			}
			assert.Equal(c.exp1, received1)
			assert.Equal(c.exp2, received2)
		})
	}
}

func TestTimerReset(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	clock := clockwork.NewFakeClock()
	alarm1 := clock.Now().Add(1 * time.Second)
	alarm2 := clock.Now().Add(2 * time.Second)
	duration := clock.Now().Add(2 * time.Second)

	user := tgapi.User{Id: 1}

	var received bool
	engine := engine.NewEngineMock()
	engine.On(
		"Receive",
		mock.Anything,
		mock.Anything,
	).
		Return(nil).
		Run(func(args mock.Arguments) { received = true })

	timer := newTimer(ctx, engine, clock, time.Second)
	timer.SetAlarm(user, "1", "1", alarm1)
	timer.SetAlarm(user, "1", "1", alarm2)

	for clock.Now().Before(duration) {
		assert.False(received)
		timer.advance(time.Second)
	}
	assert.True(received)
}
