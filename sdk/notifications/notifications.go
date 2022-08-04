package notifications

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	telegramClient "github.com/pranavsindura/at-watch/connections/telegram"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	"github.com/pranavsindura/at-watch/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NotifyPositionalCanEnter(strategyID primitive.ObjectID, userID primitive.ObjectID, tradeType int, candleClose float64) error {
	chatID, err := telegramHelpers.GetChatIDByUserID(userID)
	if err != nil {
		return err
	}
	text := "Waiting to Enter\n\n"
	text += "Strategy: " + strategyConstants.StrategyPositional + "\n"
	text += "Strategy ID: " + strategyID.Hex() + "\n"
	text += "Trade Type: " + marketConstants.TradeTypeToTextMap[tradeType] + "\n"
	text += "Candle Close: " + utils.RoundFloat(candleClose) + "\n"
	msg := tgbotapi.NewMessage(chatID, text)
	telegramClient.Client().Send(msg)

	return nil
}

func NotifyPositionalCanExit(strategyID primitive.ObjectID, userID primitive.ObjectID, exitReason int, candleClose float64, PL float64) error {
	chatID, err := telegramHelpers.GetChatIDByUserID(userID)
	if err != nil {
		return err
	}
	text := "Waiting to Exit\n\n"
	text += "Strategy: " + strategyConstants.StrategyPositional + "\n"
	text += "Strategy ID: " + strategyID.Hex() + "\n"
	text += "Exit Reason: " + marketConstants.TradeExitReasonToTextMap[exitReason] + "\n"
	text += "Candle Close: " + utils.RoundFloat(candleClose) + "\n"
	text += "PL: " + utils.RoundFloat(PL) + "\n"
	msg := tgbotapi.NewMessage(chatID, text)
	telegramClient.Client().Send(msg)

	return nil
}

func NotifyCreators(text string) {

}
