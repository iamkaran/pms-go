// Package broker is responsible for recieving MQTT PUB/SUB from the thingsboard gateway
// It is also responsible for routing the RPC and Attribute update requests from Thingsboard to the thingsboard gateways
package broker

import (
	"log/slog"

	"github.com/iamkaran/pms-go/internal/config"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func ServerMQTT(brokerCfg config.BrokerConfig, hookCfg config.BrokerHookConfig, topicCfg config.TopicList, log *slog.Logger) (chan TelemetryMsg, chan CriticalMsg, func(), error) {
	server := mqtt.New(&mqtt.Options{})

	stop := func() {
		err := server.Close()
		if err != nil {
			return
		}
	}

	if hookCfg.AllowAny {
		// To allow any connections
		if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
			log.Error("add hook", "error", err)
			return nil, nil, nil, err
		}
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      brokerCfg.TCPID,
		Address: brokerCfg.Address,
	})

	err := server.AddListener(tcp)
	if err != nil {
		log.Error("tcp listener", "error", err)
		return nil, nil, nil, err
	}

	telemetryChan := make(chan TelemetryMsg, 100)
	criticalChan := make(chan CriticalMsg, 100)

	err = server.AddHook(&GatewayHooks{
		logger:        log,
		brokerCfg:     brokerCfg,
		hookCfg:       hookCfg,
		telemetryChan: telemetryChan,
		criticalChan:  criticalChan,
		topicsCfg:     topicCfg,
	}, nil)
	if err != nil {
		log.Error("add hook", "error", err)
		return nil, nil, nil, err
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Error("serve", "error", err)
		}
	}()

	return telemetryChan, criticalChan, stop, nil
}
