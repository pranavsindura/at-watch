package telegramCommands

import (
	"strconv"
	"time"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Ping(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	receiveTime := update.Message.Time()
	sendTime := time.Now()
	diff := sendTime.UnixMilli() - receiveTime.UnixMilli()
	msgText := "pong!🏓\n\ntook " + strconv.Itoa(int(diff)) + " ms"
	msg := telegramBot.NewMessage(update.Message.Chat.ID, msgText)
	bot.Send(msg)
	return nil
}
