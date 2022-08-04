package telegramCommands

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// func generateResumeStrategyUsage() string {
// 	return "Invalid arguments, please use this format\n\n/" + telegramConstants.CommandResumeStrategy + `<strategy> <strategyID>`
// }

func resumeStrategy(update tgbotapi.Update, strategy string, strategyIDHex string) (*tgbotapi.MessageConfig, error) {
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
		// update the isActive of the strategy requested by our user
		res, err := positionalStrategyModel.
			GetPositionalStrategyCollection().
			UpdateOne(context.Background(), bson.M{
				"_id":    strategyID,
				"userID": userID,
			}, bson.M{
				"$set": bson.M{
					"isActive": true,
				},
			})

		if err != nil {
			return nil, err
		}
		if res.MatchedCount == 0 {
			return nil, fmt.Errorf("no such strategy was found")
		}
		if res.ModifiedCount == 0 {
			return telegramUtils.GenerateReplyMessage(update, "Strategy is already resumed"), nil
		}
		return telegramUtils.GenerateReplyMessage(update, "Resumed Strategy "+strategy+" "+strategyIDHex), nil
	default:
		return nil, fmt.Errorf("handling for this strategy does not exist")
	}
}

func ResumeStrategy(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
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

	msg, err := resumeStrategy(update, strategy, strategyIDHex)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
