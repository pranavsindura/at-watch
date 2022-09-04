package telegramCommands

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/constants"
	"github.com/pranavsindura/at-watch/crons"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func stopMarket(update telegramBot.Update) (*telegramBot.MessageConfig, error) {
	_, err := marketSDK.Stop()

	if err != nil {
		return nil, err
	}
	crons.MarketCron().Stop()
	notifications.Broadcast(constants.AccessLevelUser, "Market has now Stopped")

	return telegramUtils.GenerateReplyMessage(update, "Successfully Stopped Market"), nil
}

func StopMarket(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	msg, err := stopMarket(update)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
