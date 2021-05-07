package main

import (
	"fmt"
	"log"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

const BOT_TOKEN = "1005504135:AAFxLo6xN-2nVRm_AUJJdJWNEaNJLc6WmAE"
const CHAT_ID = 147524383 // sam personal chat
// const CHAT_ID := 18446744073403423083 // real roommate chat

var messengerBot *tg.BotAPI

func listenForUpdates() {
	updatesBot, err := tg.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	updateConfig := tg.NewUpdate(0)
	updates, err := updatesBot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Listening for Telegram commands")

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			return
		}

		log.Printf("Received command %s", update.Message.Command())

		args := strings.Split(update.Message.CommandArguments(), " ")
		resp := tg.NewMessage(update.Message.Chat.ID, "")

		validMachines := strings.Join(machineNames, ", ")

		if len(args) != 1 {
			resp.Text = fmt.Sprintf("Please provide a machine: %s", validMachines)
			updatesBot.Send(resp)
			continue
		}

		var machine *LaundryMachine
		var ok bool

		if machine, ok = machines[args[0]]; !ok {
			resp.Text = fmt.Sprintf("Invalid machine name %s, valid ones are: %s", args[0], validMachines)
			updatesBot.Send(resp)
			continue
		}

		var roommate *Roommate

		for _, r := range roommates {
			if r.ChatId == update.Message.Chat.ID {
				roommate = r
				break
			}
		}

		if roommate == nil {
			resp.Text = fmt.Sprintf("User %s with ID %d was not found", update.Message.Chat.UserName, update.Message.Chat.ID)
			updatesBot.Send(resp)
			continue
		}

		switch update.Message.Command() {
		case "claim":
			err := machine.Claim(roommate)
			if err != nil {
				resp.Text = err.Error()
			} else {
				resp.Text = fmt.Sprintf("%s claimed by %s", machine.GetName(), roommate.Name)
			}

		case "unclaim":
			err := machine.Unclaim(roommate)
			if err != nil {
				resp.Text = err.Error()
			} else {
				resp.Text = fmt.Sprintf("%s unclaimed by %s", machine.GetName(), roommate.Name)
			}

		case "collect":
			machine.MarkCollected()
			resp.Text = fmt.Sprintf("%s was collected by %s", machine.GetName(), roommate.Name)

		default:
			resp.Text = fmt.Sprintf("Unknown command %s. Valid commands are: /claim, /unclaim, /collect", args[0])
		}

		updatesBot.Send(resp)
	}
}

func init() {
	var err error

	messengerBot, err = tg.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	go listenForUpdates()
}

func SendMessageToAll(message string) {
	sendMessage(CHAT_ID, message)
}

func sendMessage(chatId int64, message string) {
	log.Printf("Sending Telegram message %s to %d", message, chatId)
	msg := tg.NewMessage(chatId, message)
	messengerBot.Send(msg)
}
