package messaging

import (
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/terabyte128/golaundrybot/machine"
)

const BOT_TOKEN = "1005504135:AAFxLo6xN-2nVRm_AUJJdJWNEaNJLc6WmAE"
const CHAT_ID = 147524383 // sam personal chat
// const CHAT_ID := 18446744073403423083 // real roommate chat

var bot *tg.BotAPI
var updates tg.UpdatesChannel

func listenForUpdates() {
	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			return
		}

		resp := tg.NewMessage(update.Message.Chat.ID, "")

		var selectedMachine machine.LaundryMachine

		for mac := range main.Machines {

		}

		if update.Message.Command() == "claim" {

		}
	}
}

func init() {
	bot, err := tg.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	updateConfig := tg.NewUpdate(0)
	updates, err = bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal(err)
	}

	go listenForUpdates()
}

func SendMessageToAll(message string) {
	sendMessage(CHAT_ID, message)
}

func sendMessage(chatId int64, message string) {
	msg := tg.NewMessage(chatId, message)
	bot.Send(msg)
}
