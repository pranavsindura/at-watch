package telegramMiddlewares

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/constants"
	telegramConstants "github.com/pranavsindura/at-watch/constants/telegram"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func AccessLevelHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string) error {
	if _, exists := telegramConstants.MinimumAccessLevel[command]; !exists {
		return fmt.Errorf("unable to determine access level for this command")
	}
	requiredAccessLevel := telegramConstants.MinimumAccessLevel[command]
	if requiredAccessLevel == constants.AccessLevelCustom {
		return nil
	}
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return err
	}

	if userSession.AccessLevel < telegramConstants.MinimumAccessLevel[command] {
		return telegramUtils.GenerateMinimumAccessLevelError(userSession.AccessLevel, telegramConstants.MinimumAccessLevel[command])
	}

	return nil
}

func FyersAccessTokenExists(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string) error {
	fyersAccessToken := fyersSDK.GetFyersAccessToken()
	if fyersAccessToken == "" {
		return fmt.Errorf("fyers access token does not exist")
	}

	return nil
}

func MarketNotActiveAndNotWarmingUp(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string) error {
	if marketSDK.IsMarketWatchActive() || marketSDK.IsWarmUpInProgress() {
		return fmt.Errorf("this operation is not allowed during market hours")
	}

	return nil
}

func MarketNotWarmingUp(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string) error {
	if marketSDK.IsWarmUpInProgress() {
		return fmt.Errorf("this operation is not allowed while market is warming up")
	}

	return nil
}

// func MarketInactive(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string) error {
// 	if !marketSDK.IsMarketWatchActive() && !marketSDK.IsWarmUpInProgress() {
// 		return fmt.Errorf("this operation is not allowed during market hours")
// 	}

// 	return nil
// }

func ErrorHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string, err error) {
	bot.Send(telegramUtils.GenerateReplyMessage(update, telegramUtils.GenerateGenericErrorText(err)))
}
