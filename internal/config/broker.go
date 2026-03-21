package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type BrokerConfig struct {
	TCPID   string `yaml:"tcp_id"`
	Address string `yaml:"address"`
}

type BrokerHookConfig struct {
	HookID string `yaml:"hook_id"`
}

type brokerFile struct {
	Broker BrokerConfig     `yaml:"broker"`
	Hook   BrokerHookConfig `yaml:"hook"`
}

func loadBrokerConfig(cfg *BrokerConfig, path string) error {
	var b brokerFile
	if err := cleanenv.ReadConfig(path, &b); err != nil {
		return fmt.Errorf("loading broker config: %w", err)
	}
	*cfg = b.Broker
	return nil
}
