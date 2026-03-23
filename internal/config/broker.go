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
	HookID   string `yaml:"hook_id"`
	AllowAny bool   `yaml:"allow_any"`
}

type brokerFile struct {
	Broker BrokerConfig     `yaml:"broker"`
	Hook   BrokerHookConfig `yaml:"hooks"`
}

func loadBrokerConfig(brokerCfg *BrokerConfig, hookCfg *BrokerHookConfig, path string) error {
	var b brokerFile
	if err := cleanenv.ReadConfig(path, &b); err != nil {
		return fmt.Errorf("loading broker config: %w", err)
	}
	*brokerCfg = b.Broker
	*hookCfg = b.Hook
	fmt.Printf("allow any: %v", b.Hook.AllowAny)
	return b.Validate()
}

func (b *brokerFile) Validate() error {
	if b.Hook.AllowAny {
		fmt.Println("[WARNING] broker is using insecure option <allow any connection>")
		return nil // For development phase only
	}
	return nil
}
