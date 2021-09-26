package timer

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"golang.org/x/xerrors"

	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

type Config struct {
	Period time.Duration `yaml:"period"`
}

type timerKey struct {
	Type string
	Name string
}

type TimerEvent struct {
	Type     string
	Name     string
	Receiver tgapi.User
	Time     time.Time
}

func (t *TimerEvent) key() timerKey { return timerKey{Type: t.Type, Name: t.Name} }

func (t *TimerEvent) User() tgapi.User                                              { return t.Receiver }
func (t *TimerEvent) Message() interface{}                                          { return t }
func (t *TimerEvent) PreProcess(ctx context.Context, client *tgapi.TGClient) error  { return nil }
func (t *TimerEvent) PostProcess(ctx context.Context, client *tgapi.TGClient) error { return nil }

type Timer struct {
	mx     sync.Mutex
	events map[tgapi.User]map[timerKey]time.Time
	queue  []*TimerEvent

	ticker clockwork.Ticker
}

func NewTimer(ctx context.Context, cfg Config, eng *engine.Engine) *Timer {
	return newTimer(ctx, eng, clockwork.NewRealClock().NewTicker(cfg.Period))
}

func NewFakeTimer(ctx context.Context, eng *engine.Engine, clock clockwork.Clock, period time.Duration) *Timer {
	return newTimer(ctx, eng, clock.NewTicker(period))
}

func newTimer(ctx context.Context, eng *engine.Engine, ticker clockwork.Ticker) *Timer {
	res := &Timer{
		events: map[tgapi.User]map[timerKey]time.Time{},
		ticker: ticker,
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-res.ticker.Chan():
				now := time.Now()
				res.mx.Lock()
				i := sort.Search(len(res.queue), func(i int) bool { return res.queue[i].Time.After(now) })
				process := res.queue[:i]
				res.queue = res.queue[i:]
				res.mx.Unlock()
				for _, event := range process {
					err := eng.Receive(ctx, event)
					if xerrors.Is(err, engine.BadStateError) || xerrors.Is(err, engine.RetriableError) {
						// retry it next time
						res.mx.Lock()
						res.queue = append(res.queue, event)
						res.mx.Unlock()
						continue
					}
					res.mx.Lock()
					delete(res.events[event.Receiver], event.key())
					res.mx.Unlock()
				}
			}
		}
	}()
	return res
}

func (t *Timer) SetAlarm(user tgapi.User, name string, typ string, at time.Time) {
	t.mx.Lock()
	defer t.mx.Unlock()

	if _, ok := t.events[user]; !ok {
		t.events[user] = map[timerKey]time.Time{}
	}
	if old, ok := t.events[user][timerKey{typ, name}]; ok {
		if old.Equal(at) {
			return
		}
		// replace
		t.events[user][timerKey{typ, name}] = at
		i := sort.Search(len(t.queue), func(i int) bool { return !t.queue[i].Time.Before(old) })
		t.queue[i].Time = at
	} else {
		t.queue = append(t.queue, &TimerEvent{Name: name, Type: typ, Receiver: user, Time: at})
	}
	sort.Slice(t.queue, func(i, j int) bool { return t.queue[i].Time.Before(t.queue[i].Time) })
}
