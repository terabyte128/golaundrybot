package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

var BOT_TOKEN = ""
var CHAT_ID int64 = 0

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

		valid := false
		for _, command := range [...]string{"claim", "unclaim", "collect"} {
			if update.Message.Command() == command {
				valid = true
			}
		}

		if !valid {
			continue
		}

		var roommate *Roommate

		for _, r := range roommates {
			if r.ChatId == update.Message.Chat.ID {
				roommate = r
				break
			}
		}

		resp := tg.NewMessage(update.Message.Chat.ID, "")

		if roommate == nil {
			resp.Text = fmt.Sprintf("User %s with ID %d was not found", update.Message.Chat.UserName, update.Message.Chat.ID)
			updatesBot.Send(resp)
			continue
		}

		args := strings.Split(update.Message.CommandArguments(), " ")
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

		switch update.Message.Command() {
		case "claim":
			machine.Claim(roommate)
			resp.Text = fmt.Sprintf("%s claimed by %s", machine.GetName(), roommate.Name)

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

	if token, ok := os.LookupEnv("BOT_TOKEN"); ok {
		BOT_TOKEN = token
	} else {
		log.Fatal("Missing BOT_TOKEN")
	}

	if roommate_chat, ok := os.LookupEnv("ROOMMATE_CHAT_ID"); ok {
		CHAT_ID, err = strconv.ParseInt(roommate_chat, 10, 64)

		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("Missing ROOMMATE_CHAT_ID")
	}

	for _, roommate := range roommates {
		varName := fmt.Sprintf("%s_CHAT_ID", strings.ToUpper(roommate.Name))

		if chatId, ok := os.LookupEnv(varName); ok {
			roommate.ChatId, err = strconv.ParseInt(chatId, 10, 64)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("Missing %s", varName)
		}
	}

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
