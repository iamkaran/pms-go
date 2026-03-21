package broker

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/iamkaran/pms-go/internal/config"
	"github.com/iamkaran/pms-go/internal/logger"
	"testing"
	"time"
)

const (
	topic   = "test/broker"
	message = "this better work"
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
		errChan := make(chan error, 1)
		go func() {
			errChan <- ServerMQTT(cfg.Broker, cfg.Hook, log)
		}()
		time.Sleep(100 * time.Millisecond)

		client := createClient(cfg.Broker)
		token := client.Connect()
		token.Wait()
		if token.Error() != nil {
			t.Fatalf("failed connecting to broker: %v", token.Error())
		}

		client.Subscribe(topic, 0, messagePubHandler).Wait()
		client.Publish(topic, 0, false, message)

		select {
		case msg := <-mqttMsgChan:
			if msg != message {
				t.Fatalf("got %s, want %s", msg, message)
			}
		case <-time.After(time.Second * 2):
			t.Fatalf("timed out waiting for response ")
		case err := <-errChan:
			t.Fatalf("error starting mqtt broker: %v", err)
		}
	})
}
