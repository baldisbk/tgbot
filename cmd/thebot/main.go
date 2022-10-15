package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baldisbk/tgbot_sample/internal/config"
	"github.com/baldisbk/tgbot_sample/internal/impl"
	"github.com/baldisbk/tgbot_sample/internal/usercache"
	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/poller"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
	"go.uber.org/zap"
)

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Logger: %#v", err)
		os.Exit(1)
	}
	ctx = logging.WithLogger(ctx, logger)

	config, err := config.ParseConfig()
	if err != nil {
		logging.S(ctx).Errorf("Read config: %#v", err)
		os.Exit(1)
	}

	tgClient, err := tgapi.NewClient(ctx, config.ApiConfig)
	if err != nil {
		logging.S(ctx).Errorf("TG client: %#v", err)
		os.Exit(1)
	}

	cache, err := usercache.NewCache(config.CacheConfig)
	if err != nil {
		logging.S(ctx).Errorf("DB client: %#v", err)
		os.Exit(1)
	}
	defer cache.Close()

	eng := engine.NewEngine(tgClient, cache)

	tim := timer.NewTimer(ctx, config.TimerConfig, eng)
	defer tim.Shutdown()

	factory := impl.NewFactory(config.FactoryConfig, tgClient, tim)
	cache.AttachFactory(factory)

	poll := poller.NewPoller(ctx, config.PollerConfig, tgClient, eng)
	defer poll.Shutdown()

	<-signals
	cancel()
}
