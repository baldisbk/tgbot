package poller

import (
	"context"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

type Poller struct {
	Client tgapi.TGClient
	Engine engine.Engine

	config Config
	clock  clockwork.Clock

	wg sync.WaitGroup
}

type Config struct {
	PollPeriod time.Duration `yaml:"period"`
}

func NewPoller(ctx context.Context, cfg Config, client tgapi.TGClient, engine engine.Engine) *Poller {
	return newPoller(ctx, cfg, clockwork.NewRealClock(), client, engine)
}

func newPoller(ctx context.Context, cfg Config, clock clockwork.Clock,
	client tgapi.TGClient, engine engine.Engine) *Poller {
	poller := &Poller{
		Client: client,
		Engine: engine,
		config: cfg,
		clock:  clock,
	}
	poller.wg.Add(1)
	go poller.run(ctx)
	return poller
}

func (p *Poller) Shutdown() { p.wg.Wait() }

func (p *Poller) run(ctx context.Context) {
	ticker := p.clock.NewTicker(p.config.PollPeriod)
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.Chan():
		}
		upds, err := p.Client.GetUpdates(ctx)
		if err != nil {
			logging.S(ctx).Errorf("Error getting updates: %#v", err)
			continue
		}
		for _, upd := range upds {
			var err error
			switch {
			case upd.Message != nil:
				ctx = logging.WithTag(ctx, "EVENT", upd.Message.UUID)
				err = p.Engine.Receive(ctx, upd.Message)
			case upd.CallbackQuery != nil:
				ctx = logging.WithTag(ctx, "EVENT", upd.CallbackQuery.UUID)
				err = p.Engine.Receive(ctx, upd.CallbackQuery)
			}
			// TODO process different errors
			if err != nil {
				logging.S(ctx).Errorf("Error processing update (%#v): %#v", upd, err)
			}
		}
	}
}
