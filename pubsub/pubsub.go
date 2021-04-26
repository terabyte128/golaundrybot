package pubsub

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

var client mqtt.Client

func init() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://katara:1883")
	opts.SetClientID("golaundrybot")
	opts.SetUsername("arastra")
	opts.SetPassword("arastra")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client = mqtt.NewClient(opts)
}

func Connect() {
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func Subscribe(topic string, qos byte, cb mqtt.MessageHandler) {
	client.Subscribe(topic, qos, cb)
}

func Publish(topic string, qos byte, payload interface{}) {
	client.Publish(topic, qos, false, payload)
}
