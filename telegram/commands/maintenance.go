package telegramCommands

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/cache"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func maintenance(update telegramBot.Update) (*telegramBot.MessageConfig, error) {
	fyersSDK.SetFyersAccessToken("")
	cache.ClearAll()
	marketSDK.Stop()

	return telegramUtils.GenerateReplyMessage(update, "DONE"), nil
}

func Maintenance(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	msg, err := maintenance(update)
	if err != nil {
		return err
	}
	bot.Send(msg)
	return nil
}
