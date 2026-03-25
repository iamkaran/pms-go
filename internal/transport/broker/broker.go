// Package broker is responsible for recieving MQTT PUB/SUB from the thingsboard gateway
// It is also responsible for routing the RPC and Attribute update requests from Thingsboard to the thingsboard gateways
package broker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/iamkaran/pms-go/internal/config"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

type MQTTServerResult struct {
	TelemetryCh chan TelemetryMsg
	CriticalCh  chan CriticalMsg
	ErrorCh     chan error
	Error       error
}

type MQTTServerConfig struct {
	Broker  config.BrokerConfig
	Hook    config.BrokerHookConfig
	Topic   config.TopicList
	Log     *slog.Logger
	Address string
}

func MQTTServer(ctx context.Context, cfg MQTTServerConfig) MQTTServerResult {
	serverCtx, serverCancel := context.WithCancel(ctx)
	errCh := make(chan error, 2)

	fail := func(err error) MQTTServerResult {
		serverCancel()
		return MQTTServerResult{Error: err}
	}

	caps := mqtt.NewDefaultServerCapabilities()
	caps.MaximumSessionExpiryInterval = 3600
	caps.MaximumClientWritesPending = 1024
	caps.MaximumInflight = 100

	server := mqtt.New(&mqtt.Options{
		Capabilities: caps,
		Logger:       cfg.Log,
	})

	if cfg.Hook.AllowAny {
		// To allow any connections
		if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
			cfg.Log.Error("add hook", "error", err)
			return fail(err)
		}
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
		return fail(err)
	}

	telemetryChan := make(chan TelemetryMsg, 100)
	criticalChan := make(chan CriticalMsg, 100)

	err = server.AddHook(&GatewayHooks{
		ctx:           serverCtx,
		logger:        cfg.Log,
		brokerCfg:     cfg.Broker,
		hookCfg:       cfg.Hook,
		topicsCfg:     cfg.Topic,
		telemetryChan: telemetryChan,
		criticalChan:  criticalChan,
	}, nil)
	if err != nil {
		cfg.Log.Error("add hook", "error", err)
		return fail(err)
	}

	var once sync.Once

	safeClose := func(err error) {
		once.Do(func() {
			closingErr := server.Close()
			serverCancel()

			var finalErr error
			if err != nil {
				finalErr = err
			} else {
				finalErr = closingErr
			}

			errCh <- finalErr

			close(errCh)
		})
	}

	go func() {
		if err := server.Serve(); err != nil {
			safeClose(err)
		}
	}()
	go func() {
		<-serverCtx.Done()
		safeClose(nil)
	}()

	return MQTTServerResult{
		TelemetryCh: telemetryChan,
		CriticalCh:  criticalChan,
		ErrorCh:     errCh,
		Error:       nil,
	}
}
