package telegramCommands

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	fyersConstants "github.com/pranavsindura/at-watch/constants/fyers"
	"github.com/pranavsindura/at-watch/generator"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func adminLogin(update telegramBot.Update) *telegramBot.MessageConfig {
	loginUrl := fyersSDK.GenerateAuthCodeURL(generator.GenerateLoginState(fyersConstants.AdminTelegramUserID))
	return telegramUtils.GenerateReplyMessage(update, loginUrl)
}

func AdminLogin(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	msg := adminLogin(update)
	bot.Send(msg)
	return nil
}
