// Package config involves loading config files from the /config directory and then populating the structs of each component's config structure.
// Each component has its on *.go file in this package that contain methods for validation and loading specific to that component.
// This approach allows flexibility when adding new config components.
package config

type Config struct {
	ThingsBoard ThingsboardConfig
	Log         LoggerConfig
	Broker      BrokerConfig
	Hook        BrokerHookConfig
	Topics      TopicList
}

func Load(configDir string) (*Config, error) {
	cfg := &Config{}

	if err := loadThingsBoard(&cfg.ThingsBoard, configDir+"/thingsboard.yaml"); err != nil {
		return nil, err
	}

	if err := loadLogger(&cfg.Log, configDir+"/log.yaml"); err != nil {
		return nil, err
	}

	if err := loadBrokerConfig(&cfg.Broker, &cfg.Hook, configDir+"/broker.yaml"); err != nil {
		return nil, err
	}

	if err := loadTopicConfig(&cfg.Topics, configDir+"/topics.yaml"); err != nil {
		return nil, err
	}

	return cfg, nil
}
