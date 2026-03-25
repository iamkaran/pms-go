// Package broker is responsible for recieving MQTT PUB/SUB from the thingsboard gateway
// It is also responsible for routing the RPC and Attribute update requests from Thingsboard to the thingsboard gateways
package broker

import (
	"errors"
	"log/slog"
	"time"

	"github.com/iamkaran/pms-go/internal/config"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

type MQTTServerResult struct {
	TelemetryCh chan TelemetryMsg
	CriticalCh  chan CriticalMsg
	Shutdown    func() error
	Error       error
}

type MQTTServerConfig struct {
	Broker  config.BrokerConfig
	Hook    config.BrokerHookConfig
	Topic   config.TopicList
	Log     *slog.Logger
	Address string
}

func MQTTServer(cfg MQTTServerConfig) MQTTServerResult {
	caps := mqtt.NewDefaultServerCapabilities()
	caps.MaximumSessionExpiryInterval = 3600
	caps.MaximumClientWritesPending = 1024
	caps.MaximumInflight = 100

	server := mqtt.New(&mqtt.Options{
		Capabilities: caps,
		Logger:       cfg.Log,
	})

	stop := func() error {
		err := server.Close()
		if err != nil {
			return err
		}
		return nil
	}

	if cfg.Hook.AllowAny {
		// To allow any connections
		if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
			cfg.Log.Error("add hook", "error", err)
			return MQTTServerResult{nil, nil, nil, err}
		}
	}

	ready := make(chan struct{})
	rh := &ReadyHook{ready: ready}
	if err := server.AddHook(rh, nil); err != nil {
		return MQTTServerResult{Error: err}
	}

	brokerAddress := cfg.Address
	if brokerAddress == "" {
		brokerAddress = cfg.Broker.Address
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      cfg.Broker.TCPID,
		Address: brokerAddress,
	})

	err := server.AddListener(tcp)
	if err != nil {
		cfg.Log.Error("tcp listener", "error", err)
		return MQTTServerResult{nil, nil, nil, err}
	}

	telemetryChan := make(chan TelemetryMsg, 100)
	criticalChan := make(chan CriticalMsg, 100)

	err = server.AddHook(&GatewayHooks{
		logger:        cfg.Log,
		brokerCfg:     cfg.Broker,
		hookCfg:       cfg.Hook,
		topicsCfg:     cfg.Topic,
		telemetryChan: telemetryChan,
		criticalChan:  criticalChan,
	}, nil)
	if err != nil {
		cfg.Log.Error("add hook", "error", err)
		return MQTTServerResult{nil, nil, nil, err}
	}

	go func() {
		err := server.Serve()
		if err != nil {
			cfg.Log.Error("serve", "error", err)
		}
	}()

	select {
	case <-rh.ready:
		cfg.Log.Info("server started", "result", "success")
	case <-time.After(5 * time.Second):
		return MQTTServerResult{Error: errors.New("server start timed out")}
	}

	return MQTTServerResult{telemetryChan, criticalChan, stop, nil}
}
