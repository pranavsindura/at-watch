package telegramCommands

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/generator"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func login(update telegramBot.Update) *telegramBot.MessageConfig {
	telegramUserID := update.Message.From.ID

	loginUrl := fyersSDK.GenerateAuthCodeURL(generator.GenerateLoginState(telegramUserID))
	return telegramUtils.GenerateReplyMessage(update, loginUrl)
}

func Login(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	msg := login(update)
	bot.Send(msg)
	return nil
}
