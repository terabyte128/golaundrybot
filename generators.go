package main

import (
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/terabyte128/golaundrybot/machine"
	"github.com/terabyte128/golaundrybot/messaging"
)

func generateMachineUpdateFn(mac *machine.LaundryMachine) func(client mqtt.Client, message mqtt.Message) {
	return func(client mqtt.Client, message mqtt.Message) {
		amps, err := strconv.ParseFloat(string(message.Payload()), 32)

		if err != nil {
			log.Fatal(err)
		}

		mac.Update(float32(amps))
		mac.NotifyUser()
		publishStates()
	}
}

func generateButtonUpdateFn(mac *machine.LaundryMachine, roommate *messaging.Roommate) func(client mqtt.Client, message mqtt.Message) {
	return func(client mqtt.Client, message mqtt.Message) {
		mac.ButtonPress(roommate)
		publishStates()
	}
}
