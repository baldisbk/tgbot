package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/baldisbk/tgbot_sample/internal/config"
	"github.com/baldisbk/tgbot_sample/internal/tgmock"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
)

func main() {
	var err error
	fmt.Printf("Starting mock...\n")

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

	var cfg tgmock.Config
	flags, err := config.ParseCustomConfig(&cfg)
	if err != nil {
		logging.S(ctx).Errorf("Read config: %#v", err)
		os.Exit(1)
	}
	cfg.ConfigFlags = *flags

	server := tgmock.NewServer(ctx, cfg)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.S(ctx).Errorf("Serve error: %s", err)
			os.Exit(1)
		}
	}()
	defer server.Shutdown(ctx)

	<-signals
}
