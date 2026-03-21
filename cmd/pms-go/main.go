package main

import (
	"fmt"
	"github.com/iamkaran/pms-go/internal/config"
	"github.com/iamkaran/pms-go/internal/logger"
	"github.com/iamkaran/pms-go/internal/transport/embedded-broker"
	"os"
)

func main() {
	cfg, err := config.Load("config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
	}
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("pms-go starting")

	err = embeddedbroker.MqttBroker(log)
	if err != nil {
		log.Error("error starting broker", "error", err)
	}
}
