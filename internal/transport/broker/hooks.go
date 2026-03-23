package broker

import (
	"log/slog"
	"time"

	"github.com/iamkaran/pms-go/internal/config"
	"github.com/iamkaran/pms-go/internal/core"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type CriticalMsg struct {
	Type       core.EventType
	Topic      string
	Payload    []byte
	RecievedAt time.Time
}

type TelemetryMsg struct {
	Topic   string
	Payload []byte
}

type GatewayHooks struct {
	mqtt.HookBase
	logger        *slog.Logger
	brokerCfg     config.BrokerConfig
	hookCfg       config.BrokerHookConfig
	topicsCfg     config.TopicList
	criticalChan  chan CriticalMsg
	telemetryChan chan TelemetryMsg
}

func (gh *GatewayHooks) ID() string {
	return gh.hookCfg.HookID
}

func (gh *GatewayHooks) Provides(b byte) bool {
	return b == mqtt.OnPublish || b == mqtt.OnConnect
}

func (gh *GatewayHooks) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	gh.logger.Info("tb-gateway connected", "client_id", cl.ID, "topic", pk.TopicName)
	return nil
}

func (gh *GatewayHooks) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	switch pk.TopicName {
	case gh.topicsCfg.TelemetryTopic:
		gh.echoEvent(gh.topicsCfg.TelemetryTopic, cl, pk)
	case gh.topicsCfg.AttributeTopic:
		gh.echoEvent(gh.topicsCfg.AttributeTopic, cl, pk)
	case gh.topicsCfg.ConnectTopic:
		gh.echoEvent(gh.topicsCfg.ConnectTopic, cl, pk)
	case gh.topicsCfg.DisconnectTopic:
		gh.echoEvent(gh.topicsCfg.DisconnectTopic, cl, pk)
	}
	return pk, nil
}

func (gh *GatewayHooks) echoEvent(topic string, cl *mqtt.Client, pk packets.Packet) {
	gh.logger.Info("message recieved", "topic", topic, "client_id", cl.ID, "payload", string(pk.Payload))
	gh.telemetryChan <- TelemetryMsg{Topic: pk.TopicName, Payload: pk.Payload}
}
