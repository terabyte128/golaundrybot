package machine

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/terabyte128/golaundrybot/messaging"
)

// light behavior constants. these need to be explicitly defined because they are used externally
const (
	LIGHT_FAST_BLINK = 1
	LIGHT_SLOW_BLINK = 2
	LIGHT_ON         = 3
	LIGHT_OFF        = 4
)

// laundry machine states
const (
	STATE_READY              = iota // no one using and no button pressed
	STATE_CLAIMED            = iota // button pressed but not yet on
	STATE_RUNNING            = iota // running
	STATE_WAITING_COLLECTION = iota // finished but not yet collected
)

var statePatterns = map[int]int{
	STATE_READY:              LIGHT_ON,
	STATE_CLAIMED:            LIGHT_SLOW_BLINK,
	STATE_RUNNING:            LIGHT_SLOW_BLINK,
	STATE_WAITING_COLLECTION: LIGHT_FAST_BLINK,
}

type LaundryMachine struct {
	name           string              // name for debugging purposes
	ampThreshold   float32             // threshold for when considered on
	idleTimeoutSec int                 // how long to wait before assuming off
	lastOnTime     time.Time           // when the machine was last seen on
	currentState   int                 // current state from state enum
	user           *messaging.Roommate // current user
}

// NewLaundryMachine creates a new machine
func NewLaundryMachine(name string, ampThreshold float32, idleTimeoutSec int) *LaundryMachine {
	return &LaundryMachine{
		name:           name,
		ampThreshold:   ampThreshold,
		idleTimeoutSec: idleTimeoutSec,
		lastOnTime:     time.Unix(0, 0),
		currentState:   STATE_READY,
	}
}

// Update updates the machine's state with the current amp reading and return the new state
func (machine *LaundryMachine) Update(ampReading float32) int {
	if ampReading > machine.ampThreshold {
		if machine.currentState != STATE_RUNNING {
			log.Printf("%s turned on", machine.name)
		}
		machine.lastOnTime = time.Now()
		machine.currentState = STATE_RUNNING
	} else if time.Since(machine.lastOnTime).Seconds() > float64(machine.idleTimeoutSec) {
		if machine.currentState == STATE_RUNNING {
			log.Printf("%s turned off", machine.name)
			machine.currentState = STATE_WAITING_COLLECTION
		}
	}

	return machine.currentState
}

// Handle what happens when someone presses their button for this machine
func (machine *LaundryMachine) ButtonPress(user *messaging.Roommate) {
	if machine.currentState != STATE_WAITING_COLLECTION {
		if machine.user != nil && machine.user == user {
			// reset it
			log.Printf("%s resetting user", machine.name)
			machine.user = nil
		} else if machine.user == nil {
			log.Printf("%s setting user to %s", machine.name, user.Name)
			machine.user = user

			if machine.currentState == STATE_READY {
				machine.currentState = STATE_CLAIMED
			}
		}
	} else if machine.currentState == STATE_WAITING_COLLECTION {
		// mark as done and collected
		log.Printf("%s marking laundry collected for %v", machine.name, user)
		machine.MarkCollected()
	}
}

func (machine *LaundryMachine) Claim(user *messaging.Roommate) error {
	if machine.user == nil {
		machine.user = user
		machine.currentState = STATE_CLAIMED
		return nil
	} else {
		return fmt.Errorf("load has already been claimed by %s", machine.user.Name)
	}
}

func (machine *LaundryMachine) Unclaim(user *messaging.Roommate) error {
	if machine.user == nil {
		return errors.New("this machine has not been claimed")
	} else if machine.user != user {
		return fmt.Errorf("%s has claimed this machine, not you", machine.user.Name)
	}

	machine.user = nil
	return nil
}

func (machine *LaundryMachine) NotifyUser() {
	message := fmt.Sprintf("The %s is finished, come get your laundry.", machine.name)

	if machine.currentState != STATE_WAITING_COLLECTION {
		return
	}

	if machine.user != nil {
		machine.user.SendMessage(message)
	} else {
		messaging.SendMessageToAll(message)
	}
}

func (machine *LaundryMachine) MarkCollected() {
	machine.currentState = STATE_READY
	machine.user = nil
}

func (machine *LaundryMachine) SetUser(user *messaging.Roommate) {
	log.Printf("%s set user to %s", machine.name, machine.user.Name)
	machine.user = user
}

func (machine *LaundryMachine) GetUser() *messaging.Roommate {
	return machine.user
}

// GetState returns the machine's current state
func (machine *LaundryMachine) GetState() int {
	return machine.currentState
}

func (machine *LaundryMachine) GetName() string {
	return machine.name
}

// Get light states for each roommate's button
func (machine *LaundryMachine) GetLightStates() map[*messaging.Roommate]int {
	roommateMap := make(map[*messaging.Roommate]int)

	if machine.user != nil {
		// just turn on that button
		for _, roommate := range messaging.Roommates {
			roommateMap[roommate] = LIGHT_OFF
		}
		roommateMap[machine.user] = statePatterns[machine.currentState]
	} else {
		// turn on all the buttons
		for _, roommate := range messaging.Roommates {
			roommateMap[roommate] = statePatterns[machine.currentState]
		}
	}

	return roommateMap
}
