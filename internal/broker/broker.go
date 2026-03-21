package broker

import (
	mqtt "github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/listeners"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func MqttBroker(log *slog.Logger) error {
	server := mqtt.NewServer(nil)
	tcp := listeners.NewTCP("mt1", ":1883")

	err := server.AddListener(tcp, nil)
	if err != nil {
		log.Error("AddListener", "error", err)
		return err
	}
	err = server.Serve()
	if err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return server.Close()
}
