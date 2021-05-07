package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var machines = map[string]*LaundryMachine{
	"washer": NewLaundryMachine("washer", 3.5, 10),
	"dryer":  NewLaundryMachine("dryer", 8, 10),
}

var machineNames = make([]string, len(machines))

func publishStates() {
	for _, machine := range machines {
		MqttPublish(fmt.Sprintf("garage/laundry/%s/machineState", machine.GetName()), 1, fmt.Sprint(machine.GetState()))

		for roommate, lightState := range machine.GetLightStates() {
			fmt.Printf("publishing states %v: %v\n", roommate, lightState)
			MqttPublish(
				fmt.Sprintf("garage/laundry/%s/buttons/%s/lightState", machine.GetName(), roommate.Name), 1,
				fmt.Sprint(lightState),
			)
		}
	}
}

func main() {
	MqttConnect()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	i := 0

	for _, machine := range machines {
		// subscribe to machine updates
		MqttSubscribe(
			fmt.Sprintf("garage/laundry/%s/ampReading", machine.GetName()), 1,
			generateMachineUpdateFn(machine),
		)

		machineNames[i] = machine.GetName()
		i++

		for _, roommate := range roommates {
			// subscribe to button updates
			MqttSubscribe(
				fmt.Sprintf("garage/laundry/%s/buttons/%s/pressed", machine.GetName(), roommate.Name), 1,
				generateButtonUpdateFn(machine, roommate),
			)
		}
	}

	sig := <-done
	log.Printf("Received %s, shutting down.", sig.String())
}
