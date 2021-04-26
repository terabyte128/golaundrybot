package main

import (
	"fmt"

	"github.com/terabyte128/golaundrybot/machine"
	"github.com/terabyte128/golaundrybot/messaging"
	"github.com/terabyte128/golaundrybot/pubsub"
)

var Machines = []*machine.LaundryMachine{
	machine.NewLaundryMachine("washer", 3.5, 10),
	machine.NewLaundryMachine("dryer", 8, 10),
}

func publishStates() {
	for _, machine := range machines {
		pubsub.Publish(fmt.Sprintf("garage/laundry/%s/machineState", machine.GetName()), 1, fmt.Sprint(machine.GetState()))

		for roommate, lightState := range machine.GetLightStates() {
			fmt.Printf("publishing states %v: %v\n", roommate, lightState)
			pubsub.Publish(
				fmt.Sprintf("garage/laundry/%s/buttons/%s/lightState", machine.GetName(), roommate.Name), 1,
				fmt.Sprint(lightState),
			)
		}
	}
}

func main() {
	pubsub.Connect()

	for _, machine := range machines {
		// subscribe to machine updates
		pubsub.Subscribe(
			fmt.Sprintf("garage/laundry/%s/ampReading", machine.GetName()), 1,
			generateMachineUpdateFn(machine),
		)

		for _, roommate := range messaging.Roommates {
			// subscribe to button updates
			pubsub.Subscribe(
				fmt.Sprintf("garage/laundry/%s/buttons/%s/pressed", machine.GetName(), roommate.Name), 1,
				generateButtonUpdateFn(machine, roommate),
			)
		}
	}

	for {
	}
}
