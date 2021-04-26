package main

type Roommate struct {
	Name   string // name
	ChatId int64  // chat ID in telegram
}

var roommates = []*Roommate{
	{
		Name:   "Sam",
		ChatId: 147524383,
	},
	{
		Name:   "Claire",
		ChatId: 0,
	},
	{
		Name:   "Luke",
		ChatId: 0,
	},
	{
		Name:   "Kris",
		ChatId: 0,
	},
}

func (roommate *Roommate) SendMessage(message string) {
	sendMessage(roommate.ChatId, message)
}
