package telegramCommands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	positionalTradeModel "github.com/pranavsindura/at-watch/models/positionalTrade"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// func generateRemoveStrategyUsage() string {
// 	return "Invalid arguments, please use this format\n\n/" + telegramConstants.CommandRemoveStrategy + `<strategy> <strategyID>`
// }

func removeStrategy(update tgbotapi.Update, strategy string, strategyIDHex string) (*tgbotapi.MessageConfig, error) {
	strategyID, err := primitive.ObjectIDFromHex(strategyIDHex)
	if err != nil {
		return nil, err
	}

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
			return nil, err
		}

		// try to find open trades for this strategy
		res = positionalTradeModel.
			GetPositionalTradeCollection().
			FindOne(context.Background(), bson.M{
				"strategyID": strategyID,
				"status":     marketConstants.TradeStatusOpen,
			})
		err = res.Err()
		if err == nil {
			posObj := positionalTradeModel.PositionalTrade{}
			res.Decode(&posObj)
			// it must have found something
			return nil, fmt.Errorf("found existing open trades with this strategy, please close them first - " + posObj.ID.Hex())
		} else {
			if err != mongo.ErrNoDocuments {
				return nil, err
			}
		}
		// did not find any open trades, good to go

		// delete all positional trades with this strategy id
		delTradesResult, err := positionalTradeModel.
			GetPositionalTradeCollection().
			DeleteMany(context.Background(), bson.M{
				"strategyID": strategyID,
			})
		if err != nil {
			return nil, err
		}

		// delete the strategy
		res = positionalStrategyModel.
			GetPositionalStrategyCollection().
			FindOneAndDelete(context.Background(), bson.M{
				"_id": strategyID,
			})
		err = res.Err()
		if err != nil {
			return nil, err
		}
		return telegramUtils.GenerateReplyMessage(update, "Removed Strategy "+strategy+" "+strategyIDHex+" and "+strconv.Itoa(int(delTradesResult.DeletedCount))+" trades"), nil
	default:
		return nil, fmt.Errorf("handling for this strategy does not exist")
	}
}

func RemoveStrategy(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	argString := update.Message.CommandArguments()
	argList := strings.Split(argString, " ")

	if len(argList) != 2 {
		return fmt.Errorf("invalid arguments")
	}

	strategy := argList[0]
	strategyIDHex := argList[1]

	if _, exists := strategyConstants.Strategies[strategy]; !exists {
		return fmt.Errorf("invalid strategy")
	}

	msg, err := removeStrategy(update, strategy, strategyIDHex)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
