package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

const pollPeriod = time.Second

type Poller struct {
	Client *tgapi.TGClient
	Engine *engine.Engine
}

func NewPoller(ctx context.Context, client *tgapi.TGClient, engine *engine.Engine) *Poller {
	poller := &Poller{
		Client: client,
		Engine: engine,
	}
	go poller.run(ctx)
	return poller
}

func (p *Poller) run(ctx context.Context) {
	ticker := time.NewTicker(pollPeriod)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		upds, err := p.Client.GetUpdates(ctx)
		if err != nil {
			fmt.Printf("Error getting updates: %#v\n", err)
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
				fmt.Printf("Error processing update (%#v): %#v\n", upd, err)
			}
		}
	}
}
