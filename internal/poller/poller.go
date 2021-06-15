package poller

import (
	"fmt"
	"time"

	"github.com/baldisbk/tgbot_sample/internal/engine"
	"github.com/baldisbk/tgbot_sample/internal/tgapi"
)

const pollPeriod = time.Second

type Poller struct {
	Client *tgapi.TGClient
	Engine *engine.Engine

	stopper chan struct{}
}

func NewPoller(client *tgapi.TGClient, engine *engine.Engine) *Poller {
	return &Poller{
		Client:  client,
		Engine:  engine,
		stopper: make(chan struct{}),
	}
}

func (p *Poller) Run() {
	ticker := time.NewTicker(pollPeriod)
	for {
		select {
		case <-p.stopper:
			return
		case <-ticker.C:
		}
		upds, err := p.Client.GetUpdates()
		if err != nil {
			fmt.Printf("Error getting updates: %#v\n", err)
			continue
		}
		for _, upd := range upds {
			var err error
			switch {
			case upd.Message != nil:
				err = p.Engine.Receive(upd.Message)
			case upd.CallbackQuery != nil:
				err = p.Engine.Receive(upd.CallbackQuery)
			}
			if err != nil {
				fmt.Printf("Error processing update (%#v): %#v\n", upd, err)
			}
		}
	}
}

func (p *Poller) Close() {
	close(p.stopper)
}
