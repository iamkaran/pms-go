package broker

import (
	"context"
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
	ReceivedAt time.Time
}

type TelemetryMsg struct {
	Topic   string
	Payload []byte
}

// TODO: Add a waitgroup to prevent hanging writers

type GatewayHooks struct {
	mqtt.HookBase
	ctx           context.Context
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
	gh.echoEvent(pk.TopicName, cl, pk)
	return pk, nil
}

func (gh *GatewayHooks) echoEvent(topic string, cl *mqtt.Client, pk packets.Packet) {
	gh.logger.Info("message received", "topic", topic, "client_id", cl.ID, "payload", string(pk.Payload))

	switch topic {
	case gh.topicsCfg.TelemetryTopic:
		select {
		case gh.telemetryChan <- TelemetryMsg{Topic: pk.TopicName, Payload: pk.Payload}:
		case <-gh.ctx.Done():
			gh.logger.Warn("context cancelled, dropping telemetry msg", "topic", topic)
		default:
			gh.logger.Warn("channel full, dropping message", "topic", topic)
		}
	case gh.topicsCfg.AttributeTopic,
		gh.topicsCfg.ConnectTopic,
		gh.topicsCfg.DisconnectTopic:

		var eventType core.EventType
		switch topic {
		case gh.topicsCfg.AttributeTopic:
			eventType = core.EventAttribute
		case gh.topicsCfg.ConnectTopic:
			eventType = core.EventConnect
		case gh.topicsCfg.DisconnectTopic:
			eventType = core.EventDisconnect
		}

		select {
		case gh.criticalChan <- CriticalMsg{
			Type:       eventType,
			Topic:      pk.TopicName,
			Payload:    pk.Payload,
			ReceivedAt: time.Now(),
		}:
		case <-gh.ctx.Done():
			// TODO: Improve the handling of dropped messages
			gh.logger.Error("context cancelled, dropping critical msg", "topic", topic)
		}
	}
}
