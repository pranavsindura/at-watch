package telegramCommands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// func generateEnterStrategyUsage() string {
// 	return "invalid arguments, please use this format\n\n/" + telegramConstants.CommandEnterStrategy + `
// <strategy>
// <strategyID>
// <tradeType>
// <price>
// <lots>`
// }

func enterStrategy(update tgbotapi.Update, strategy string, strategyID primitive.ObjectID, tradeType int, price float64, lots int) (*tgbotapi.MessageConfig, error) {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return nil, err
	}

	userID := userSession.UserID

	switch strategy {
	case strategyConstants.StrategyPositional:
		// find this strategy and if it was requested by the user
		res := positionalStrategyModel.
			GetPositionalStrategyCollection().
			FindOne(context.Background(), bson.M{
				"_id":    strategyID,
				"userID": userID,
			})
		err := res.Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, fmt.Errorf("no strategy with this strategyID was found")
			}
			return nil, err
		}
		// must have found something
		_, err = marketSDK.EnterPositionalTrade(strategyID, price, lots, tradeType, time.Now())

		if err != nil {
			return nil, err
		}

		return telegramUtils.GenerateReplyMessage(update, "Successfully entered Trade"), nil
	default:
		return nil, fmt.Errorf("handling for this strategy does not exist")
	}
}

func EnterStrategy(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	argString := update.Message.CommandArguments()
	argList := strings.Split(argString, "\n")
	if len(argList) != 5 {
		return fmt.Errorf("invalid arguments")
	}

	strategy := argList[0]
	strategyIDHex := argList[1]
	tradeTypeText := argList[2]
	priceString := argList[3]
	lotsString := argList[4]

	if _, exists := strategyConstants.Strategies[strategy]; !exists {
		return fmt.Errorf("invalid strategy")
	}

	strategyID, err := primitive.ObjectIDFromHex(strategyIDHex)
	if err != nil {
		return err
	}

	if _, exists := marketConstants.TradeTypeTextToTypeMap[tradeTypeText]; !exists {
		return fmt.Errorf("invalid tradeType")
	}
	tradeType := marketConstants.TradeTypeTextToTypeMap[tradeTypeText]

	price, err := strconv.ParseFloat(priceString, 64)
	if err != nil {
		return err
	}

	lots, err := strconv.ParseInt(lotsString, 10, 64)
	if err != nil {
		return err
	}

	msg, err := enterStrategy(update, strategy, strategyID, tradeType, price, int(lots))

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
