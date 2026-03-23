package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type TopicList struct {
	TelemetryTopic  string `yaml:"telemetry"`
	AttributeTopic  string `yaml:"attribute"`
	ConnectTopic    string `yaml:"connect"`
	DisconnectTopic string `yaml:"disconnect"`
}

type topicFile struct {
	Topics TopicList `yaml:"topics"`
}

func loadTopicConfig(cfg *TopicList, path string) error {
	var t topicFile
	if err := cleanenv.ReadConfig(path, &t); err != nil {
		return fmt.Errorf("loading topics config: %w", err)
	}
	*cfg = t.Topics
	return t.Validate()
}

func (t *topicFile) Validate() error {
	topics := map[string]string{
		"telemetry":  t.Topics.TelemetryTopic,
		"attribute":  t.Topics.AttributeTopic,
		"connect":    t.Topics.ConnectTopic,
		"disconnect": t.Topics.DisconnectTopic,
	}
	for name, value := range topics {
		if value == "" {
			return fmt.Errorf("topic name cannot be empty for %s", name)
		}
	}
	return nil
}
