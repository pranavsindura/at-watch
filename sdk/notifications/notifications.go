package notifications

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	telegramClient "github.com/pranavsindura/at-watch/connections/telegram"
	"github.com/pranavsindura/at-watch/constants"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	"github.com/pranavsindura/at-watch/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Notify(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := telegramClient.Client().Send(msg)
	return err
}

func NotifyPositionalCanEnter(strategyID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, tradeType int, candleClose float64) error {
	chatID, err := telegramHelpers.GetChatIDByUserID(userID)
	if err != nil {
		return err
	}
	text := "Waiting to Enter\n\n"
	text += "Strategy: " + strategyConstants.StrategyPositional + "\n"
	text += "Strategy ID: " + strategyID.Hex() + "\n"
	text += "Trade Type: " + marketConstants.TradeTypeToTextMap[tradeType] + "\n"
	text += "Instrument: " + instrument + "\n"
	text += "Time Frame: " + marketConstants.TimeFrameToTextMap[timeFrame] + "\n"
	text += "Candle Close: " + utils.RoundFloat(candleClose) + "\n"
	err = Notify(chatID, text)
	if err != nil {
		return err
	}
	return nil
}

func NotifyPositionalCanExit(strategyID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, exitReason int, candleClose float64, PL float64) error {
	chatID, err := telegramHelpers.GetChatIDByUserID(userID)
	if err != nil {
		return err
	}
	text := "Waiting to Exit\n\n"
	text += "Strategy: " + strategyConstants.StrategyPositional + "\n"
	text += "Strategy ID: " + strategyID.Hex() + "\n"
	text += "Exit Reason: " + marketConstants.TradeExitReasonToTextMap[exitReason] + "\n"
	text += "Instrument: " + instrument + "\n"
	text += "Time Frame: " + marketConstants.TimeFrameToTextMap[timeFrame] + "\n"
	text += "Candle Close: " + utils.RoundFloat(candleClose) + "\n"
	text += "PL: " + utils.RoundFloat(PL) + "\n"
	err = Notify(chatID, text)
	if err != nil {
		return err
	}

	return nil
}

func Broadcast(accessLevel int, text string) error {
	serverBroadcastText := "[SERVER BROADCAST][" + constants.AccessLevelToTextMap[accessLevel] + "]\n\n" + text
	res, err := telegramUserModel.
		GetTelegramUserCollection().
		Find(context.Background(), bson.M{
			"accessLevel": bson.M{
				"$gte": accessLevel,
			},
		})
	if err != nil {
		return err
	}

	for res.Next(context.Background()) {
		user := telegramUserModel.TelegramUserModel{}
		res.Decode(&user)
		err := Notify(user.TelegramChatID, serverBroadcastText)
		if err != nil {
			fmt.Println("unable to send message to user", err)
		}
	}

	return nil
}
