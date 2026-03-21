package embeddedbroker

import (
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func MqttBroker(log *slog.Logger) error {
	server := mqtt.New(&mqtt.Options{})

	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		return err
	}

	if err := server.AddHook(&GatewayHook{logger: log}, nil); err != nil {
		return err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: ":1883",
	})

	if err := server.AddListener(tcp); err != nil {
		log.Error("failed to add tcp listener", "error", err)
		return err
	}

	if err := server.Serve(); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	return server.Close()
}
