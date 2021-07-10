package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baldisbk/tgbot_sample/internal/impl"
	"github.com/baldisbk/tgbot_sample/internal/usercache"
	"github.com/baldisbk/tgbot_sample/pkg/engine"
	"github.com/baldisbk/tgbot_sample/pkg/poller"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/baldisbk/tgbot_sample/pkg/timer"
)

const dbName = "db.sqlite"

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	tgClient, err := tgapi.NewClient(ctx, tgapi.TgApi, tgapi.BotToken)
	if err != nil {
		fmt.Printf("TG client: %#v", err)
		os.Exit(1)
	}

	cache, err := usercache.NewCache(usercache.Config{Filename: dbName})
	if err != nil {
		fmt.Printf("DB client: %#v", err)
		os.Exit(1)
	}
	defer cache.Close()

	eng := engine.NewEngine(tgClient, cache)

	tim := timer.NewTimer(ctx, eng)

	factory := impl.NewFactory(tgClient, tim)
	cache.AttachFactory(factory)

	poller.NewPoller(ctx, tgClient, eng)

	<-signals
}
