package config

type Config struct {
	ThingsBoard ThingsboardConfig
	Log         LoggerConfig
}

func Load(configDir string) (*Config, error) {
	cfg := &Config{}

	if err := loadThingsBoard(&cfg.ThingsBoard, configDir+"/thingsboard.yaml"); err != nil {
		return nil, err
	}

	if err := loadLogger(&cfg.Log, configDir+"/log.yaml"); err != nil {
		return nil, err
	}

	return cfg, nil
}
