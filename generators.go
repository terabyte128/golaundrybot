package main

import (
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func generateMachineUpdateFn(mac *LaundryMachine) func(client mqtt.Client, message mqtt.Message) {
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

func generateButtonUpdateFn(mac *LaundryMachine, roommate *Roommate) func(client mqtt.Client, message mqtt.Message) {
	return func(client mqtt.Client, message mqtt.Message) {
		mac.ButtonPress(roommate)
		publishStates()
	}
}
