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
	case "v1/gateway/telemetry":
		gh.logger.Info("<telemetry> recieved", "client_id", cl.ID, "payload", string(pk.Payload))
		gh.telemetryChan <- TelemetryMsg{Topic: pk.TopicName, Payload: pk.Payload}
	case "v1/gateway/connect":
		gh.logger.Info("<connect> recieved", "client_id", cl.ID, "payload", string(pk.Payload))
		gh.criticalChan <- CriticalMsg{Type: "connect", Topic: pk.TopicName, Payload: pk.Payload, RecievedAt: time.Now()}
	case "v1/gateway/disconnect":
		gh.logger.Info("<disconnect> recieved", "client_id", cl.ID, "payload", string(pk.Payload))
		gh.criticalChan <- CriticalMsg{Type: "disconnect", Topic: pk.TopicName, Payload: pk.Payload, RecievedAt: time.Now()}
	}
	return pk, nil
}
