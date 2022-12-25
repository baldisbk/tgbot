package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/baldisbk/tgbot_sample/internal/config"
	"github.com/baldisbk/tgbot_sample/internal/impl"
	"github.com/baldisbk/tgbot_sample/internal/usercache"
	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/poller"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
)

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Logger: %#v", err)
		os.Exit(1)
	}
	defer logger.Sync()
	ctx = logging.WithLogger(ctx, logger)

	logging.S(ctx).Debugf("Starting bot...")

	logging.S(ctx).Debugf("Parsing config...")

	config, err := config.ParseConfig()
	if err != nil {
		logging.S(ctx).Errorf("Read config: %#v", err)
		os.Exit(1)
	}
	logging.S(ctx).Debugf("Config: %#v", config)

	logging.S(ctx).Debugf("Init TG client...")

	tgClient, err := tgapi.NewClient(ctx, config.ApiConfig)
	if err != nil {
		logging.S(ctx).Errorf("TG client: %#v", err)
		os.Exit(1)
	}

	logging.S(ctx).Debugf("Init database...")

	cache, err := usercache.NewCache(ctx, config.CacheConfig)
	if err != nil {
		logging.S(ctx).Errorf("DB client: %#v", err)
		os.Exit(1)
	}
	defer cache.Close()

	eng := engine.NewEngine(tgClient, cache)

	logging.S(ctx).Debugf("Starting timers...")

	tim := timer.NewTimer(ctx, config.TimerConfig, eng)
	defer tim.Shutdown()

	factory := impl.NewFactory(config.FactoryConfig, tgClient, tim)
	if err := cache.AttachFactory(ctx, factory); err != nil {
		logging.S(ctx).Errorf("attach factory: %#v", err)
		os.Exit(1)
	}

	logging.S(ctx).Debugf("Starting poller...")

	poll := poller.NewPoller(ctx, config.PollerConfig, tgClient, eng)
	defer poll.Shutdown()

	logging.S(ctx).Debugf("Bot started")

	<-signals
}
