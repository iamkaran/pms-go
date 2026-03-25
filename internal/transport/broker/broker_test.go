package broker

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/iamkaran/pms-go/internal/config"
	"github.com/iamkaran/pms-go/internal/logger"
)

const (
	message    = "Test Message"
	configPath = "../../../config"
)

func makeMessageHandler(ch chan string) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		ch <- string(msg.Payload())
	}
}

func createClient(address string, clientID string) mqtt.Client {
	broker := fmt.Sprintf("tcp://127.0.0.1%s", address)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)

	return mqtt.NewClient(opts)
}

func connectClient(t testing.TB, client mqtt.Client) {
	token := client.Connect()
	token.Wait()
	err := token.Error()
	if err != nil {
		t.Fatalf("error connecting to broker: %v", err)
	}
}

func publishMessage(msgChan chan string, publishTopic string, client mqtt.Client, publishMessage string) {
	client.Subscribe(publishTopic, 0, makeMessageHandler(msgChan)).Wait()
	client.Publish(publishTopic, 0, false, publishMessage).Wait()
}

func createMQTTServer(t testing.TB, ctx context.Context, cfg *config.Config, log *slog.Logger, address string) MQTTServerResult {
	t.Helper()
	serverResult := MQTTServer(ctx, MQTTServerConfig{
		Broker:  cfg.Broker,
		Hook:    cfg.Hook,
		Topic:   cfg.Topics,
		Log:     log,
		Address: address,
	})

	if serverResult.Error != nil {
		t.Fatalf("error serving mqtt broker: %v", serverResult.Error)
	}
	return serverResult
}

func TestServeMQTT(t *testing.T) {
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("error loading config %v", err)
	}

	log := logger.New("info", "json")

	t.Run("test pub and sub", func(t *testing.T) {
		address := ":1883"
		ctx, cancel := context.WithCancel(context.Background())
		serverResult := createMQTTServer(t, ctx, cfg, log, address)

		client := createClient(address, "go-test-client-1")
		connectClient(t, client)
		defer client.Disconnect(250)

		msgChan := make(chan string, 1)
		publishMessage(msgChan, "test/topic", client, message)

		select {
		case msg := <-msgChan:
			if msg != message {
				t.Fatalf("got %s, want %s", msg, message)
			}
		case <-time.After(time.Second * 2):
			t.Fatalf("timed out waiting for response")
		case err := <-serverResult.ErrorCh:
			t.Fatalf("server internal error: %v", err)
		}
		cancel()
		err = <-serverResult.ErrorCh
		if err != nil {
			log.Error("server close", "error", err)
		} else {
			log.Info("server close", "status", "success")
		}
	})
	t.Run("test telemetry hooking mechanism of broker", func(t *testing.T) {
		address := ":1883"
		ctx, cancel := context.WithCancel(context.Background())
		serverResult := createMQTTServer(t, ctx, cfg, log, address)

		client := createClient(address, "go-test-client-1")
		connectClient(t, client)
		defer client.Disconnect(250)

		msgChan := make(chan string, 10)
		publishMessage(msgChan, cfg.Topics.TelemetryTopic, client, message)

		select {
		case msg := <-serverResult.TelemetryCh:
			t.Logf("telemetry hook triggered, topic :%s", string(msg.Topic))
		case <-time.After(time.Second * 2):
			t.Fatalf("timed out waiting for response ")
		case err := <-serverResult.ErrorCh:
			t.Fatalf("server internal error: %v", err)
		}

		criticalTopics := []string{
			cfg.Topics.AttributeTopic,
			cfg.Topics.ConnectTopic,
			cfg.Topics.DisconnectTopic,
		}

		for _, topic := range criticalTopics {
			publishMessage(msgChan, topic, client, message)
			select {
			case msg := <-serverResult.CriticalCh:
				t.Logf("type critical hook triggered, topic :%s", string(msg.Topic))
			case <-time.After(time.Second * 2):
				t.Fatalf("timed out waiting for response ")
			case err := <-serverResult.ErrorCh:
				t.Fatalf("server internal error: %v", err)
			}
		}

		cancel()

		err = <-serverResult.ErrorCh
		if err != nil {
			log.Error("server close", "error", err)
		} else {
			log.Info("server close", "status", "success")
		}
	})
}
