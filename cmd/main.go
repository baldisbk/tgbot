package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/baldisbk/tgbot_sample/internal/engine"
	"github.com/baldisbk/tgbot_sample/internal/impl"
	"github.com/baldisbk/tgbot_sample/internal/poller"
	"github.com/baldisbk/tgbot_sample/internal/tgapi"
	"github.com/baldisbk/tgbot_sample/internal/usercache"
)

const dbName = "db.sqlite"

func main() {
	var err error
	var wg sync.WaitGroup
	defer wg.Wait()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	tgClient, err := tgapi.NewClient(tgapi.TgApi, tgapi.BotToken)
	if err != nil {
		fmt.Printf("TG client: %#v", err)
		os.Exit(1)
	}

	factory := impl.NewFactory(tgClient)

	cache, err := usercache.NewCache(usercache.Config{Filename: dbName}, factory.Factory)
	if err != nil {
		fmt.Printf("DB client: %#v", err)
		os.Exit(1)
	}
	defer cache.Close()

	eng := engine.NewEngine(tgClient, cache)

	poller := poller.NewPoller(tgClient, eng)
	wg.Add(1)
	go func() {
		poller.Run()
		wg.Done()
	}()
	defer poller.Close()

	<-signals
}
