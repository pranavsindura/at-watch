package telegramCommands

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getOpenTrade(update tgbotapi.Update, userID primitive.ObjectID) (*tgbotapi.MessageConfig, error) {
	text := ""
	openTradeCount := 0
	// Positional
	posOpenTrades, posStrategies, err := marketSDK.GetPositionalOpenTrades(userID)
	if err != nil {
		return nil, err
	}

	openTradeCount += len(posOpenTrades)
	if len(posOpenTrades) > 0 {
		text += strategyConstants.StrategyPositional + "\n"
	}

	for idx, openTrade := range posOpenTrades {
		text += "Strategy ID: " + posStrategies[idx].ID.Hex() + "\n"
		text += "Instrument: " + posStrategies[idx].Instrument + "\n"
		text += "Trade Type: " + openTrade.TradeTypeText + "\n"
		text += "Entry At: " + openTrade.Entry.Candle.DateString + "\n"
		text += "Entry Price: " + utils.RoundFloat(openTrade.Entry.Candle.Close) + "\n"
		text += "Lots: " + strconv.Itoa(openTrade.Lots) + "\n"
		text += "PL: " + utils.RoundFloat(openTrade.PL) + "\n"
		text += "LTP: " + utils.RoundFloat(openTrade.UpdatedAtLTP) + "\n"
		text += "Updated At: " + utils.GetDateStringFromTimestamp(openTrade.UpdatedAtTS) + "\n"
		text += "\n"
	}

	text = strconv.Itoa(openTradeCount) + " Open Trades\n\n" + text

	return telegramUtils.GenerateReplyMessage(update, text), nil
}

func GetOpenTrade(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return err
	}

	userID := userSession.UserID

	msg, err := getOpenTrade(update, userID)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
