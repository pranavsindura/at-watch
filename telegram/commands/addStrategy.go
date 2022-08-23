package telegramCommands

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	positionalStrategyConstants "github.com/pranavsindura/at-watch/constants/strategies/positional"
	instrumentModel "github.com/pranavsindura/at-watch/models/instrument"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
)

// func generateAddStrategyUsage() string {
// 	return "Invalid arguments, please use this format\n\n/" + telegramConstants.CommandAddStrategy + `
// 	<strategy>
// 	<instrument>
// 	<timeFrame>`
// }

func addStrategy(update tgbotapi.Update, strategy string, instrument string) (*tgbotapi.MessageConfig, error) {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return nil, err
	}

	userID := userSession.UserID

	res := instrumentModel.GetInstrumentCollection().FindOne(context.Background(), bson.M{"symbol": instrument})

	err = res.Err()
	if err != nil {
		return nil, err
	}

	instrumentObj := instrumentModel.InstrumentModel{}
	err = res.Decode(&instrumentObj)
	if err != nil {
		return nil, err
	}

	switch strategy {
	case strategyConstants.StrategyPositional:
		positionalStrategy := positionalStrategyModel.PositionalStrategy{
			UserID:       userID,
			InstrumentID: instrumentObj.ID,
			TimeFrame:    marketConstants.TimeFrameToTextMap[positionalStrategyConstants.TimeFrame],
			IsActive:     true,
		}
		coll := positionalStrategyModel.GetPositionalStrategyCollection()
		_, err := coll.InsertOne(context.Background(), positionalStrategy)
		if err != nil {
			return nil, err
		}
		return telegramUtils.GenerateReplyMessage(update, "Added Strategy "+strategy), nil
	default:
		return nil, fmt.Errorf("handling for this strategy does not exist")
	}
}

func AddStrategy(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	argString := update.Message.CommandArguments()
	argList := strings.Split(argString, "\n")
	fmt.Println(argString)
	fmt.Println(argList)
	if len(argList) != 2 {
		return fmt.Errorf("invalid arguments")
	}

	strategy := argList[0]
	instrument := argList[1]

	fmt.Println(strategy, instrument)

	instruments, err := telegramHelpers.GetInstruments()

	if err != nil {
		return err
	}

	isInstrumentValid := false
	for _, realInstrument := range instruments {
		isInstrumentValid = isInstrumentValid || instrument == realInstrument
	}

	if _, exists := strategyConstants.Strategies[strategy]; strategy == "" || !exists {
		return fmt.Errorf("invalid strategy")
	}
	if !isInstrumentValid {
		return fmt.Errorf("invalid instrument")
	}

	msg, err := addStrategy(update, strategy, instrument)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
