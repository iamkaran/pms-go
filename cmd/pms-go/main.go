package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iamkaran/pms-go/internal/config"
	"github.com/iamkaran/pms-go/internal/logger"
	"github.com/iamkaran/pms-go/internal/transport/broker"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Load("config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("pms-go starting")

	ctx, cancel := context.WithCancel(context.Background())

	log.Info("allow any", "status", cfg.Hook.AllowAny)
	serverResult := broker.MQTTServer(ctx, broker.MQTTServerConfig{
		Broker:  cfg.Broker,
		Hook:    cfg.Hook,
		Topic:   cfg.Topics,
		Log:     log,
		Address: cfg.Broker.Address,
	})
	if serverResult.Error != nil {
		log.Error("error starting broker", "error", err)
		os.Exit(1)
	}

	select {
	case <-quit:
		log.Info("shutting down")
	case err := <-serverResult.ErrorCh:
		log.Error("server nil error", "error", err)
	}

	cancel()

	err = <-serverResult.ErrorCh
	if err != nil {
		log.Error("server close", "error", err)
	} else {
		log.Info("server close", "status", "success")
	}
}
