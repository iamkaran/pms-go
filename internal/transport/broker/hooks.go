package broker

import (
	"log/slog"
	"strings"

	"github.com/iamkaran/pms-go/internal/config"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type GatewayHooks struct {
	mqtt.HookBase
	logger    *slog.Logger
	brokerCfg config.BrokerConfig
	hookCfg   config.BrokerHookConfig
}

func (gh *GatewayHooks) ID() string {
	return gh.hookCfg.HookID
}

func (gh *GatewayHooks) Provides(b byte) bool {
	return b == mqtt.OnPublish || b == mqtt.OnConnect
}

func (gh *GatewayHooks) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	gh.logger.Info("tb-gateway connected", "client_id", cl.ID)
	return nil
}

func (gh *GatewayHooks) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	if strings.Contains(pk.TopicName, "v1/gateway/") {
		gh.logger.Info("tb-gateway message recieved",
			"client_id", cl.ID,
			"payload", pk.Payload,
		)
	}
	return pk, nil
}
