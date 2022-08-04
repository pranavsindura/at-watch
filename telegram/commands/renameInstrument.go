package telegramCommands

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/cache"
	instrumentModel "github.com/pranavsindura/at-watch/models/instrument"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
)

// func generateRenameInstrumentUsage() string {
// 	return "Invalid arguments, please use this format\n\n/" + telegramConstants.CommandRenameInstrument + " <oldInstrument> <newInstrument>"
// }

func renameinstrument(update tgbotapi.Update, oldInstrument string, newInstrument string) (*tgbotapi.MessageConfig, error) {
	isValid, err := fyersSDK.IsValidInstrument(newInstrument)

	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, fmt.Errorf("invalid newInstrument")
	}

	coll := instrumentModel.GetInstrumentCollection()
	res := coll.FindOneAndUpdate(context.Background(),
		bson.M{
			"symbol": oldInstrument,
		}, bson.M{
			"$set": bson.M{
				"symbol": newInstrument,
			},
		})

	err = res.Err()

	if err != nil {
		return nil, err
	}

	cache.DeleteInstruments()

	return telegramUtils.GenerateReplyMessage(update, "Renamed Instrument "+oldInstrument+" to "+newInstrument), nil
}

func RenameInstrument(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	argString := update.Message.CommandArguments()
	argList := strings.Split(argString, " ")

	if len(argList) != 2 {
		return fmt.Errorf("invalid arguments")
	}

	oldInstrument := argList[0]
	newInstrument := argList[1]

	msg, err := renameinstrument(update, oldInstrument, newInstrument)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
