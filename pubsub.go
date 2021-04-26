package main

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("MQTT connected to broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Fatalf("Connect lost: %v", err)
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

func MqttConnect() {
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func MqttSubscribe(topic string, qos byte, cb mqtt.MessageHandler) {
	client.Subscribe(topic, qos, cb)
}

func MqttPublish(topic string, qos byte, payload interface{}) {
	client.Publish(topic, qos, false, payload)
}
