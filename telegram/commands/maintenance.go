package telegramCommands

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/cache"
	"github.com/pranavsindura/at-watch/constants"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func maintenance(update telegramBot.Update) (*telegramBot.MessageConfig, error) {
	notifications.Broadcast(constants.AccessLevelUser, "Server is performing maintenance, please avoid any actions till maintenance is over")

	fyersSDK.SetFyersAccessToken("")
	cache.ClearAll()
	marketSDK.Stop()

	notifications.Broadcast(constants.AccessLevelUser, "Server has finished maintenance")

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
