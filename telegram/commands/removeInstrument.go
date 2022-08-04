package telegramCommands

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/cache"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	instrumentModel "github.com/pranavsindura/at-watch/models/instrument"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func removeinstrument(update tgbotapi.Update, instrument string) (*tgbotapi.MessageConfig, error) {
	coll := instrumentModel.GetInstrumentCollection()
	res := coll.FindOne(context.Background(), bson.M{
		"symbol": instrument,
	})

	err := res.Err()
	if err != nil {
		return nil, err
	}

	instrumentObj := instrumentModel.InstrumentModel{}
	res.Decode(&instrumentObj)

	instrumentID := instrumentObj.ID

	pos := positionalStrategyModel.GetPositionalStrategyCollection()
	res = pos.FindOne(context.Background(), bson.M{
		"instrumentID": instrumentID,
	})

	err = res.Err()
	if err == nil {
		posObj := positionalStrategyModel.PositionalStrategy{}
		res.Decode(&posObj)
		// it must have found something
		return nil, fmt.Errorf("found existing " + strategyConstants.StrategyPositional + " strategies with this instrument, please remove them first - " + posObj.ID.Hex())
	} else {
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
	}
	// i got mongo.ErrNoDocuments
	// i am good to go

	res = coll.FindOneAndDelete(context.Background(), bson.M{
		"symbol": instrument,
	})

	err = res.Err()
	if err != nil {
		return nil, err
	}

	cache.DeleteInstruments()

	return telegramUtils.GenerateReplyMessage(update, "Removed Instrument "+instrument), nil
}

func RemoveInstrument(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	instrument := update.Message.CommandArguments()

	msg, err := removeinstrument(update, instrument)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
