package main

import (
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
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("pms-go starting")

	log.Info("allow any", "status", cfg.Hook.AllowAny)
	_, _, stop, err := broker.ServerMQTT(cfg.Broker, cfg.Hook, cfg.Topics, log)
	if err != nil {
		log.Error("error starting broker", "error", err)
	}

	<-quit
	stop()
}
