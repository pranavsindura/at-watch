package telegramCommands

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/constants"
	mongoConstants "github.com/pranavsindura/at-watch/constants/mongo"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getStrategy(update tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return nil, err
	}

	userID := userSession.UserID

	text := ""
	strategyCount := 0

	type PositionalStrategyAggregateResult struct {
		positionalStrategyModel.PositionalStrategy `bson:",inline"`
		Instrument                                 struct {
			InstrumentID primitive.ObjectID `bson:"_id"`
			Symbol       string             `bson:"symbol"`
		} `bson:"instrument"`
	}
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "userID", Value: userID}}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: mongoConstants.InstrumentCollection}, {Key: "localField", Value: "instrumentID"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "instrument"}}}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$instrument"}}

	positionalStrategyCursor, err := positionalStrategyModel.
		GetPositionalStrategyCollection().
		Aggregate(
			context.Background(),
			mongo.Pipeline{
				matchStage,
				lookupStage,
				unwindStage,
			},
		)

	if err != nil {
		return nil, err
	}

	positionalStrategies := make([]PositionalStrategyAggregateResult, 0)
	for positionalStrategyCursor.Next(context.Background()) {
		strategy := PositionalStrategyAggregateResult{PositionalStrategy: positionalStrategyModel.PositionalStrategy{}}
		positionalStrategyCursor.Decode(&strategy)
		fmt.Println(utils.BruteStringify(strategy))
		positionalStrategies = append(positionalStrategies, strategy)
	}

	strategyCount += len(positionalStrategies)

	if len(positionalStrategies) > 0 {
		for _, strategy := range positionalStrategies {
			isActiveText := constants.No
			if strategy.IsActive {
				isActiveText = constants.Yes
			}
			fmt.Println(isActiveText)
			text += "Strategy: " + strategyConstants.StrategyPositional + "\n"
			text += "Strategy ID: " + strategy.ID.Hex() + "\n"
			text += "Instrument: " + strategy.Instrument.Symbol + "\n"
			text += "Time Frame: " + strategy.TimeFrame + "\n"
			text += "Is Active?: " + isActiveText + "\n"
			text += "\n"
		}
	}

	text = strconv.Itoa(strategyCount) + " Strategies\n\n" + text

	return telegramUtils.GenerateReplyMessage(update, text), nil
}

func GetStrategy(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	msg, err := getStrategy(update)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
