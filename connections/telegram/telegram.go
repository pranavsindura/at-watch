package telegramClient

import (
	"os"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
	"github.com/rs/zerolog/log"
)

var Bot *telegramBot.BotAPI

func Init() {
	key := os.Getenv(envConstants.TelegramBotKey)
	newBot, err := telegramBot.NewBotAPI(key)
	if err != nil {
		log.Fatal().Err(err)
		return
	}
	Bot = newBot
	// Bot.Debug = true
}

func Client() *telegramBot.BotAPI {
	if Bot == nil {
		Init()
	}
	return Bot
}
