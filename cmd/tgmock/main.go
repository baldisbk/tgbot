package main

import (
	"context"
	"fmt"
	"os"

	"github.com/baldisbk/tgbot_sample/internal/config"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"go.uber.org/zap"
)

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// var wg sync.WaitGroup
	// defer wg.Wait()

	// signals := make(chan os.Signal)
	// signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Logger: %#v", err)
		os.Exit(1)
	}
	ctx = logging.WithLogger(ctx, logger)

	var cfg Config
	flags, err := config.ParseCustomConfig(&cfg)
	if err != nil {
		logging.S(ctx).Errorf("Read config: %#v", err)
		os.Exit(1)
	}
	cfg.ConfigFlags = *flags

	err = NewServer(ctx, cfg).ListenAndServe()
	if err != nil {
		logging.S(ctx).Errorf("serve error: %s", err)
		os.Exit(1)
	}

	// <-signals
}
