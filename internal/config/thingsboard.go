package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type ThingsboardConfig struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	UseTLS bool   `yaml:"use_tls"`
}

type thingsboardFile struct {
	Thingsboard ThingsboardConfig `yaml:"thingsboard"`
}

func loadThingsBoard(cfg *ThingsboardConfig, path string) error {
	var f thingsboardFile
	if err := cleanenv.ReadConfig(path, &f); err != nil {
		return fmt.Errorf("thingsboard config: %w", err)
	}
	*cfg = f.Thingsboard
	return cfg.Validate()
}

func (tb *ThingsboardConfig) Validate() error {
	if tb.Host == "" {
		return fmt.Errorf("thingsboard.host is required")
	}
	if tb.Port < 1 || tb.Port > 65535 {
		return fmt.Errorf("thingsboard.port must be between 1 and 65535")
	}
	return nil
}
