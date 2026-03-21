package embeddedbroker

import (
	"bytes"
	"log/slog"
	"strings"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type GatewayHook struct {
	mqtt.HookBase
	logger *slog.Logger
}

func (h *GatewayHook) ID() string {
	return "gateway-hook"
}

func (h *GatewayHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
		mqtt.OnConnect,
		mqtt.OnDisconnect,
	}, []byte{b})
}

func (h *GatewayHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	if strings.HasPrefix(pk.TopicName, "v1/gateway/") {
		h.logger.Info("Gateway message recieved",
			"topic", pk.TopicName,
			"client_id", string(cl.ID),
			"payload", pk.Payload,
		)
	}
	return pk, nil
}

func (h *GatewayHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	h.logger.Info("tb-gateway connected", "client_id", string(cl.ID))
	return nil
}

func (h *GatewayHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	h.logger.Info("tb-gateway disconnected", "client_id", string(cl.ID))
}
