package main

import (
	"fmt"
	"github.com/iamkaran/pms-go/internal/config"
	"github.com/iamkaran/pms-go/internal/logger"
	"os"
)

func main() {
	cfg, err := config.Load("config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
	}
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("pms-go starting")
}
