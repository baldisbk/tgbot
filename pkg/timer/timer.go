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

func (t *TimerEvent) User() tgapi.User                                             { return t.Receiver }
func (t *TimerEvent) Message() interface{}                                         { return t }
func (t *TimerEvent) PreProcess(ctx context.Context, client tgapi.TGClient) error  { return nil }
func (t *TimerEvent) PostProcess(ctx context.Context, client tgapi.TGClient) error { return nil }

type Timer struct {
	mx     sync.Mutex
	events map[tgapi.User]map[timerKey]time.Time
	queue  []*TimerEvent

	clock    clockwork.Clock
	wg       sync.WaitGroup
	testSync sync.WaitGroup
}

func NewTimer(ctx context.Context, cfg Config, eng engine.Engine) *Timer {
	return newTimer(ctx, eng, clockwork.NewRealClock(), cfg.Period)
}

func (t *Timer) Shutdown() {
	t.wg.Wait()
}

// warning: use this, not clock.Advance
func (t *Timer) advance(d time.Duration) {
	if c, ok := t.clock.(clockwork.FakeClock); ok {
		t.testSync.Add(1)
		c.Advance(d)
		t.testSync.Wait()
	}
}

func newTimer(ctx context.Context, eng engine.Engine, clock clockwork.Clock, period time.Duration) *Timer {
	ticker := clock.NewTicker(period)
	res := &Timer{
		events: map[tgapi.User]map[timerKey]time.Time{},
		clock:  clock,
	}
	res.wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				res.wg.Done()
				return
			case <-ticker.Chan():
				now := clock.Now()
				res.mx.Lock()
				i := sort.Search(len(res.queue), func(i int) bool { return res.queue[i].Time.After(now) })
				process := res.queue[:i]
				res.queue = res.queue[i:]
				res.mx.Unlock()

				var wg sync.WaitGroup
				wg.Add(len(process))
				for _, event := range process {
					go func(event *TimerEvent) {
						err := eng.Receive(ctx, event)
						res.mx.Lock()
						defer res.mx.Unlock()
						defer wg.Done()
						switch {
						case xerrors.Is(err, engine.BadStateError),
							xerrors.Is(err, engine.RetriableError):
							// retry it next time
							res.queue = append(res.queue, event)
						case err != nil:
							// TODO: process it somehow
						default:
							delete(res.events[event.Receiver], event.key())
						}
					}(event)
				}
				wg.Wait()
				if _, ok := res.clock.(clockwork.FakeClock); ok {
					res.testSync.Done()
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
	key := timerKey{typ, name}
	if old, ok := t.events[user][key]; ok {
		if old.Equal(at) {
			return
		}
		// replace
		t.events[user][key] = at
		i := sort.Search(len(t.queue), func(i int) bool { return !t.queue[i].Time.Before(old) })
		t.queue[i].Time = at
	} else {
		t.events[user][key] = at
		t.queue = append(t.queue, &TimerEvent{Name: name, Type: typ, Receiver: user, Time: at})
	}
	sort.Slice(t.queue, func(i, j int) bool { return t.queue[i].Time.Before(t.queue[i].Time) })
}
