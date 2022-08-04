package telegramCommands

import (
	"context"
	"fmt"
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

// func generatePauseStrategyUsage() string {
// 	return "Invalid arguments, please use this format\n\n/" + telegramConstants.CommandPauseStrategy + `<strategy> <strategyID>`
// }

func pauseStrategy(update tgbotapi.Update, strategy string, strategyIDHex string) (*tgbotapi.MessageConfig, error) {
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
		// try to find open trades for this strategy
		getRes := positionalTradeModel.
			GetPositionalTradeCollection().
			FindOne(context.Background(), bson.M{
				"strategyID": strategyID,
				"status":     marketConstants.TradeStatusOpen,
			})
		err := getRes.Err()
		if err == nil {
			posObj := positionalTradeModel.PositionalTrade{}
			getRes.Decode(&posObj)
			// it must have found something
			return nil, fmt.Errorf("found existing open trades with this strategy, please close them first - " + posObj.ID.Hex())
		} else {
			if err != mongo.ErrNoDocuments {
				return nil, err
			}
		}
		// update the isActive of the strategy requested by our user
		res, err := positionalStrategyModel.
			GetPositionalStrategyCollection().
			UpdateOne(context.Background(), bson.M{
				"_id":    strategyID,
				"userID": userID,
			}, bson.M{
				"$set": bson.M{
					"isActive": false,
				},
			})

		if err != nil {
			return nil, err
		}
		if res.MatchedCount == 0 {
			return nil, fmt.Errorf("no such strategy was found")
		}
		if res.ModifiedCount == 0 {
			return telegramUtils.GenerateReplyMessage(update, "Strategy is already paused"), nil
		}
		return telegramUtils.GenerateReplyMessage(update, "Paused Strategy "+strategy+" "+strategyIDHex), nil
	default:
		return nil, fmt.Errorf("handling for this strategy does not exist")
	}
}

func PauseStrategy(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
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

	msg, err := pauseStrategy(update, strategy, strategyIDHex)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
