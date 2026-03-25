package broker

import (
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

var (
	mqttMsgChan                           = make(chan string)
	messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		mqttMsgChan <- string(msg.Payload())
	}
)

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

func publishMessage(publishTopic string, client mqtt.Client, publishMessage string) {
	client.Subscribe(publishTopic, 0, messagePubHandler).Wait()
	client.Publish(publishTopic, 0, false, publishMessage).Wait()
}

func createMQTTServer(t testing.TB, cfg *config.Config, log *slog.Logger, address string) MQTTServerResult {
	t.Helper()
	serverResult := MQTTServer(MQTTServerConfig{
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
		serverResult := createMQTTServer(t, cfg, log, address)

		client := createClient(address, "go-test-client-1")
		connectClient(t, client)
		defer client.Disconnect(250)

		publishMessage("test/topic", client, message)

		select {
		case msg := <-mqttMsgChan:
			if msg != message {
				t.Fatalf("got %s, want %s", msg, message)
			}
		case <-time.After(time.Second * 2):
			t.Fatalf("timed out waiting for response")
		}
		if err := serverResult.Shutdown(); err != nil {
			t.Fatalf("error stopping server: %v", err)
		}
	})
	t.Run("test telemetry hooking mechanism of broker", func(t *testing.T) {
		address := ":1883"
		serverResult := createMQTTServer(t, cfg, log, address)

		client := createClient(address, "go-test-client-1")
		connectClient(t, client)
		defer client.Disconnect(250)

		publishMessage(cfg.Topics.TelemetryTopic, client, message)

		select {
		case msg := <-serverResult.TelemetryCh:
			t.Logf("telemetry hook triggered, topic :%s", string(msg.Topic))
		case <-time.After(time.Second * 2):
			t.Fatalf("timed out waiting for response ")
		}

		criticalTopics := []string{
			cfg.Topics.AttributeTopic,
			cfg.Topics.ConnectTopic,
			cfg.Topics.DisconnectTopic,
		}

		for _, topic := range criticalTopics {
			publishMessage(topic, client, message)
			select {
			case msg := <-serverResult.CriticalCh:
				t.Logf("type critical hook triggered, topic :%s", string(msg.Topic))
			case <-time.After(time.Second * 2):
				t.Fatalf("timed out waiting for response ")
			}
		}

		if err := serverResult.Shutdown(); err != nil {
			t.Fatalf("error stopping server: %v", err)
		}
	})
}
