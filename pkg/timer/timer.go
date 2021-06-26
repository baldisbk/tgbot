package timer

import (
	"sort"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

const tickerPeriod = time.Second

type TimerEvent struct {
	Name     string
	Receiver tgapi.User
	Time     time.Time
}

func (t *TimerEvent) User() tgapi.User                         { return t.Receiver }
func (t *TimerEvent) Message() interface{}                     { return t }
func (t *TimerEvent) PreProcess(client *tgapi.TGClient) error  { return nil }
func (t *TimerEvent) PostProcess(client *tgapi.TGClient) error { return nil }

type Timer struct {
	mx     sync.Mutex
	events map[tgapi.User]map[string]time.Time
	queue  []*TimerEvent

	stop   chan struct{}
	ticker clockwork.Ticker
}

func NewTimer(eng *engine.Engine) *Timer {
	return newTimer(eng, clockwork.NewRealClock().NewTicker(tickerPeriod))
}

func NewFakeTimer(eng *engine.Engine, clock clockwork.Clock, period time.Duration) *Timer {
	return newTimer(eng, clock.NewTicker(period))
}

func newTimer(eng *engine.Engine, ticker clockwork.Ticker) *Timer {
	res := &Timer{
		events: map[tgapi.User]map[string]time.Time{},
		stop:   make(chan struct{}),
		ticker: ticker,
	}
	go func() {
		for {
			select {
			case <-res.stop:
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
					// TODO process error
					_ = eng.Receive(event)
					res.mx.Lock()
					delete(res.events[event.Receiver], event.Name)
					res.mx.Unlock()
				}
			}
		}
	}()
	return res
}

func (t *Timer) Stop() { close(t.stop) }

func (t *Timer) SetAlarm(user tgapi.User, name string, at time.Time) {
	t.mx.Lock()
	defer t.mx.Unlock()

	if _, ok := t.events[user]; !ok {
		t.events[user] = map[string]time.Time{}
	}
	if old, ok := t.events[user][name]; ok {
		if old.Equal(at) {
			return
		}
		// replace
		t.events[user][name] = at
		i := sort.Search(len(t.queue), func(i int) bool { return !t.queue[i].Time.Before(old) })
		t.queue[i].Time = at
	} else {
		t.queue = append(t.queue, &TimerEvent{Name: name, Receiver: user, Time: at})
	}
	sort.Slice(t.queue, func(i, j int) bool { return t.queue[i].Time.Before(t.queue[i].Time) })
}
