package broker

import (
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/iamkaran/pms-go/internal/config"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func ServerMQTT(brokerCfg config.BrokerConfig, hookCfg config.BrokerHookConfig, log *slog.Logger) error {

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	server := mqtt.New(nil)

	// To allow any connections
	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		log.Error("add hook", "error", err)
		return err
	}

	err := server.AddHook(&GatewayHooks{logger: log, brokerCfg: brokerCfg, hookCfg: hookCfg}, nil)
	if err != nil {
		log.Error("add hook", "error", err)
		return err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      brokerCfg.TCPID,
		Address: brokerCfg.Address,
	})

	err = server.AddListener(tcp)
	if err != nil {
		log.Error("tcp listener", "error", err)
		return err
	}

	errChannel := make(chan error, 1)
	go func() {
		err := server.Serve()
		if err != nil {
			log.Error("serve", "error", err)
			errChannel <- err
		}
	}()

	<-done

	if err := server.Close(); err != nil {
		log.Error("server closing", "error", err)
		return err
	}

	return nil
}
