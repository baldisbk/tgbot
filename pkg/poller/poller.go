package poller

import (
	"context"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/baldisbk/tgbot/pkg/engine"
	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/baldisbk/tgbot/pkg/tgapi"
)

type Poller struct {
	Client tgapi.TGClient
	Engine engine.Engine

	config Config
	clock  clockwork.Clock
	offset uint64

	stopper chan struct{}
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
		Client:  client,
		Engine:  engine,
		config:  cfg,
		clock:   clock,
		stopper: make(chan struct{}),
	}
	go poller.run(ctx)
	return poller
}

func (p *Poller) Shutdown() { p.stopper <- struct{}{} }

func (p *Poller) Sync(ctx context.Context) error { return p.do(ctx, true) }

func (p *Poller) do(ctx context.Context, inSync bool) error {
	upds, offset, err := p.Client.GetUpdates(ctx, p.offset)
	if err != nil {
		return xerrors.Errorf("get updates: %w", err)
	}
	p.offset = offset
	errors := make(chan error, len(upds))
	var wg sync.WaitGroup
	if inSync {
		wg.Add(len(upds))
		go func() {
			wg.Wait()
			close(errors)
		}()
	}
	for _, upd := range upds {
		go func(upd tgapi.Update) {
			var err error
			switch {
			case upd.Message != nil:
				ctx = logging.WithTag(ctx, "EVENT", upd.Message.UUID)
				if err = p.Engine.Receive(ctx, upd.Message); err != nil {
					err = xerrors.Errorf("receive message (%#v): %w", upd.Message, err)
				}
			case upd.CallbackQuery != nil:
				ctx = logging.WithTag(ctx, "EVENT", upd.CallbackQuery.UUID)
				if err = p.Engine.Receive(ctx, upd.CallbackQuery); err != nil {
					err = xerrors.Errorf("receive callback (%#v): %w", upd.CallbackQuery, err)
				}
			}
			if err != nil {
				logging.S(ctx).Errorf("Error processing update: %#v", err)
			}
			if inSync {
				errors <- err
				wg.Done()
			}
		}(upd)
	}
	if inSync {
		var resErr error
		for err := range errors {
			if err != nil {
				resErr = multierr.Append(resErr, err)
			}
		}
		return resErr
	}
	return nil
}

func (p *Poller) run(ctx context.Context) {
	ticker := p.clock.NewTicker(p.config.PollPeriod)
	for {
		select {
		case <-ticker.Chan():
			if err := p.do(ctx, false); err != nil {
				logging.S(ctx).Errorf("Error processing updates: %#v", err)
			}
		case <-p.stopper:
			return
		case <-ctx.Done():
			return
		}
	}
}
