package telegramCommands

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/cache"
	instrumentModel "github.com/pranavsindura/at-watch/models/instrument"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

func addInstrument(update tgbotapi.Update, instrument string) (*tgbotapi.MessageConfig, error) {
	isValid, err := fyersSDK.IsValidInstrument(instrument)

	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, fmt.Errorf("invalid instrument")
	}

	coll := instrumentModel.GetInstrumentCollection()
	instrumentObj := instrumentModel.InstrumentModel{
		Symbol: instrument,
	}
	_, err = coll.InsertOne(context.Background(), instrumentObj)

	if err != nil {
		return nil, err
	}

	cache.DeleteInstruments()

	return telegramUtils.GenerateReplyMessage(update, "Added Instrument "+instrument), nil
}

func AddInstrument(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	instrument := update.Message.CommandArguments()

	msg, err := addInstrument(update, instrument)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
