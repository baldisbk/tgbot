package poller

import (
	"context"
	"time"

	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

type Poller struct {
	Client tgapi.TGClient
	Engine engine.Engine

	config Config
}

type Config struct {
	PollPeriod time.Duration `yaml:"period"`
}

func NewPoller(ctx context.Context, cfg Config, client tgapi.TGClient, engine engine.Engine) *Poller {
	poller := &Poller{
		Client: client,
		Engine: engine,
		config: cfg,
	}
	go poller.run(ctx)
	return poller
}

func (p *Poller) Shutdown() {}

func (p *Poller) run(ctx context.Context) {
	ticker := time.NewTicker(p.config.PollPeriod)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
				err = p.Engine.Receive(ctx, upd.Message)
			case upd.CallbackQuery != nil:
				err = p.Engine.Receive(ctx, upd.CallbackQuery)
			}
			// TODO process different errors
			if err != nil {
				logging.S(ctx).Errorf("Error processing update (%#v): %#v", upd, err)
			}
		}
	}
}
