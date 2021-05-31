package main

import (
	"fmt"
	"log"
	"time"
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
	Name                     string    // name for debugging purposes
	AmpThreshold             float32   // threshold for when considered on
	IdleTimeoutSec           int       // how long to wait before assuming off
	LastOnTime               time.Time // when the machine was last seen on
	LastOffTime              time.Time // when the machine was last seen off
	RunningTimeoutSec        int       // how long to wait before assuming on
	LastStartTime            time.Time // when the machine was last started
	CurrentState             int       // current state from state enum
	NextAlertTime            time.Time // when to next send an alert
	NextAlertIntervalMinutes int       // how long to wait before sending another alert once NextAlertTime is reached
	User                     *Roommate // current user
}

// NewLaundryMachine creates a new machine
func NewLaundryMachine(name string, ampThreshold float32, idleTimeoutSec int, runningTimeoutSec int) *LaundryMachine {
	return &LaundryMachine{
		Name:              name,
		AmpThreshold:      ampThreshold,
		IdleTimeoutSec:    idleTimeoutSec,
		RunningTimeoutSec: runningTimeoutSec,
		LastOnTime:        time.Unix(0, 0),
		LastOffTime:       time.Unix(0, 0),
		LastStartTime:     time.Unix(0, 0),
		CurrentState:      STATE_READY,
	}
}

// Update updates the machine's state with the current amp reading and return the new state
func (machine *LaundryMachine) Update(ampReading float32) int {
	if ampReading > machine.AmpThreshold {
		// mark RUNNING if it's been sending high values for at least RunningTimeoutSec
		if time.Since(machine.LastOffTime).Seconds() > float64(machine.RunningTimeoutSec) && machine.CurrentState != STATE_RUNNING {
			log.Printf("%s turned on", machine.Name)
			machine.LastStartTime = time.Now()
			machine.CurrentState = STATE_RUNNING

			if machine.CurrentState != STATE_CLAIMED {
				machine.User = nil // only reset when wasn't just claimed
			}
		}

		// note that we saw it on even if we don't make a transition at this time
		machine.LastOnTime = time.Now()
	} else if time.Since(machine.LastOnTime).Seconds() > float64(machine.IdleTimeoutSec) && machine.CurrentState == STATE_RUNNING {
		log.Printf("%s turned off", machine.Name)
		machine.CurrentState = STATE_WAITING_COLLECTION
		// alert after an hour, then wait 2 hours
		machine.NextAlertTime = time.Now()
		machine.NextAlertIntervalMinutes = 60
	} else {
		// off, but no state transition needed; just mark that its off
		machine.LastOffTime = time.Now()
	}

	machine.NotifyUser()
	return machine.CurrentState
}

// Handle what happens when someone presses their button for this machine
func (machine *LaundryMachine) ButtonPress(user *Roommate) {
	if machine.CurrentState != STATE_WAITING_COLLECTION {
		if machine.User != nil && machine.User == user {
			// reset it
			log.Printf("%s resetting user", machine.Name)
			machine.User = nil
		} else if machine.User == nil {
			log.Printf("%s setting user to %s", machine.Name, user.Name)
			machine.User = user

			if machine.CurrentState == STATE_READY {
				machine.CurrentState = STATE_CLAIMED
			}
		}
	} else if machine.CurrentState == STATE_WAITING_COLLECTION {
		// mark as done and collected
		log.Printf("%s marking laundry collected for %v", machine.Name, user)
		machine.MarkCollected()
	}
}

func (machine *LaundryMachine) Claim(user *Roommate) {
	machine.User = user

	if machine.CurrentState == STATE_READY {
		machine.CurrentState = STATE_CLAIMED
	}
}

func (machine *LaundryMachine) Unclaim() {
	machine.User = nil
}

func (machine *LaundryMachine) NotifyUser() {
	var message string

	if machine.CurrentState != STATE_WAITING_COLLECTION {
		return
	}

	if machine.NextAlertTime.Unix() > time.Now().Unix() {
		return
	}

	elapsedHours := int(time.Since(machine.LastOnTime).Hours())
	if elapsedHours > 0 {
		plural := "s"

		if elapsedHours == 1 {
			plural = ""
		}

		message = fmt.Sprintf("The %s has been finished for %d hour%s, come get your laundry", machine.Name, plural, elapsedHours)
	} else {
		message = fmt.Sprintf("The %s is finished, come get your laundry", machine.Name)
	}

	message = message + " and [mark it collected](http://laundry\\.wolf)\\."

	if machine.User != nil {
		machine.User.SendMessage(message)
	} else {
		message = message + "\n\nIn the future, you can [tell me this is your laundry](http://laundry\\.wolf), and I will only message you\\."
		SendMessageToAll(message)
	}

	machine.NextAlertTime = time.Now().Add(time.Minute * time.Duration(machine.NextAlertIntervalMinutes))
	machine.NextAlertIntervalMinutes *= 2

	log.Printf("Next notification will be at %v", machine.NextAlertTime)
}

func (machine *LaundryMachine) MarkCollected() {
	machine.CurrentState = STATE_READY
	machine.User = nil
}

func (machine *LaundryMachine) SetUser(user *Roommate) {
	log.Printf("%s set user to %s", machine.Name, machine.User.Name)
	machine.User = user
}

func (machine *LaundryMachine) GetUser() *Roommate {
	return machine.User
}

// GetState returns the machine's current state
func (machine *LaundryMachine) GetState() int {
	return machine.CurrentState
}

func (machine *LaundryMachine) GetName() string {
	return machine.Name
}

// Get light states for each roommate's button
func (machine *LaundryMachine) GetLightStates() map[*Roommate]int {
	roommateMap := make(map[*Roommate]int)

	if machine.User != nil {
		// just turn on that button
		for _, roommate := range roommates {
			roommateMap[roommate] = LIGHT_OFF
		}
		roommateMap[machine.User] = statePatterns[machine.CurrentState]
	} else {
		// turn on all the buttons
		for _, roommate := range roommates {
			roommateMap[roommate] = statePatterns[machine.CurrentState]
		}
	}

	return roommateMap
}

func (machine *LaundryMachine) GetFriendlyState() string {
	return [...]string{"Ready", "Claimed", "Running", "Awaiting Collection"}[machine.CurrentState]
}

func (machine *LaundryMachine) MinutesSinceStart() int {
	return int(time.Since(machine.LastStartTime).Minutes())
}

func (machine *LaundryMachine) TimeSinceStartString() string {
	minutes := machine.MinutesSinceStart()
	hours := minutes / 60
	minutes -= (60 * hours)

	out := ""
	if hours > 0 {
		out += fmt.Sprintf("%dh ", hours)
	}
	out += fmt.Sprintf("%dm", minutes)

	return out
}
