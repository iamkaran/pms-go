package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type loggerConfigFile struct {
	Log LoggerConfig `yaml:"log"`
}

func loadLogger(cfg *LoggerConfig, path string) error {
	var f loggerConfigFile
	if err := cleanenv.ReadConfig(path, &f); err != nil {
		return fmt.Errorf("log config: %w")
	}
	*cfg = f.Log
	return cfg.Validate()
}

func (l *LoggerConfig) Validate() error {
	switch l.Level {
	case "debug", "info", "error", "warn":
	default:
		return fmt.Errorf("log.Level must be a valid log level")
	}

	switch l.Format {
	case "json", "text":
	default:
		return fmt.Errorf("log.Format should be either text or json")
	}
	return nil
}
