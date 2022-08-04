package telegramCommands

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func startMarket(update telegramBot.Update) (*telegramBot.MessageConfig, error) {
	_, err := marketSDK.Start()

	if err != nil {
		return nil, err
	}

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
