package broker

import (
	"fmt"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/iamkaran/pms-go/internal/config"
	"github.com/iamkaran/pms-go/internal/logger"
)

const (
	topic          = "test/broker"
	telemetryTopic = "v1/gateway/telemetry"
	message        = "Test Message"
)

var mqttMsgChan = make(chan string)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	mqttMsgChan <- string(msg.Payload())
}

func createClient(cfg config.BrokerConfig) mqtt.Client {
	broker := fmt.Sprintf("tcp://localhost%s", cfg.Address)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("go-test-client")

	return mqtt.NewClient(opts)
}

func TestBroker(t *testing.T) {
	cfg, err := config.Load("../../../config")
	if err != nil {
		t.Fatalf("error loading config %v", err)
	}

	log := logger.New("info", "json")

	t.Run("test pub and sub", func(t *testing.T) {
		_, _, stop, err := ServerMQTT(cfg.Broker, cfg.Hook, cfg.Topics, log)
		if err != nil {
			t.Fatalf("error serving mqtt broker: %v", err)
		}
		defer stop()

		time.Sleep(100 * time.Millisecond)

		client := createClient(cfg.Broker)
		token := client.Connect()
		token.Wait()
		if token.Error() != nil {
			t.Fatalf("failed connecting to broker: %v", token.Error())
		}

		client.Subscribe(topic, 0, messagePubHandler).Wait()
		client.Publish(topic, 0, false, message).Wait()

		select {
		case msg := <-mqttMsgChan:
			if msg != message {
				t.Fatalf("got %s, want %s", msg, message)
			}
		case <-time.After(time.Second * 2):
			t.Fatalf("timed out waiting for response")
		}
	})
	t.Run("test telemetry hooking mechanism of broker", func(t *testing.T) {
		telemetryChan, _, stop, err := ServerMQTT(cfg.Broker, cfg.Hook, cfg.Topics, log)
		if err != nil {
			t.Fatalf("error serving mqtt broker: %v", err)
		}
		defer stop()

		time.Sleep(100 * time.Millisecond)

		client := createClient(cfg.Broker)
		token := client.Connect()
		token.Wait()

		if token.Error() != nil {
			t.Fatalf("failed connecting to broker: %v", token.Error())
		}

		client.Subscribe(telemetryTopic, 0, messagePubHandler).Wait()
		client.Publish(telemetryTopic, 0, false, message).Wait()

		select {
		case msg := <-telemetryChan:
			t.Logf("telemetry hook triggered, topic :%s", string(msg.Topic))
		case <-time.After(time.Second * 2):
			t.Fatalf("timed out waiting for response ")
		}
	})
}
