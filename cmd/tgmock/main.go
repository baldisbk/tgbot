package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/baldisbk/tgbot_sample/internal/config"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Config struct {
	config.ConfigFlags

	Address string `yaml:"address"`
}

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

	srv := Server{}

	mx := mux.NewRouter()
	mx.HandleFunc("/{token}/"+tgapi.ReceiveCmd, srv.update)
	mx.HandleFunc("/{token}/"+tgapi.SendCmd, srv.message)
	mx.HandleFunc("/{token}/"+tgapi.AnswerCmd, srv.callback)
	mx.HandleFunc("/{token}/"+tgapi.EditCmd, srv.message)

	mx.NotFoundHandler = http.HandlerFunc(srv.dflt)

	server := http.Server{
		Addr:        cfg.Address,
		Handler:     mx,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	err = server.ListenAndServe()
	if err != nil {
		logging.S(ctx).Errorf("serve error: %s", err)
		os.Exit(1)
	}

	// <-signals
}
