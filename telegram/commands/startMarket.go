package telegramCommands

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/constants"
	"github.com/pranavsindura/at-watch/crons"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func startMarket(update telegramBot.Update) (*telegramBot.MessageConfig, error) {
	_, err := marketSDK.Start()

	if err != nil {
		return nil, err
	}
	crons.MarketCron().Start()
	notifications.Broadcast(constants.AccessLevelUser, "Market has now Started")

	return telegramUtils.GenerateReplyMessage(update, "Successfully Started Market"), nil
}

func StartMarket(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	msg, err := startMarket(update)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
