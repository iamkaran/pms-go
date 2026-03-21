package broker

import (
	"fmt"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/iamkaran/pms-go/internal/logger"
)

const (
	mqttBroker = "tcp://localhost:1883"
	clientID   = "go-mqtt-test"
	topic      = "test-topic"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("connected to broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("connection lost: %v", err)
}

func createClient() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetClientID(clientID)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	return mqtt.NewClient(opts)
}

func TestEchoBroker(t *testing.T) {
	log := logger.New("info", "json")

	errChannel := make(chan error, 1)

	go func() {
		errChannel <- MqttBroker(log)
	}()

	time.Sleep(100 * time.Millisecond)

	client := createClient()

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	recieved := make(chan string, 1)

	client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		recieved <- string(msg.Payload())
	})

	msg := "Hello, World!"
	client.Publish(topic, 0, false, msg)

	select {
	case recievedMsg := <-recieved:
		if recievedMsg != msg {
			t.Fatalf("got %q want %q", recievedMsg, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}
}
