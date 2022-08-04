package telegramCommands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/cache"
	"github.com/pranavsindura/at-watch/constants"
	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
)

func updateAccessLevel(update tgbotapi.Update, telegramUserID int64, accessLevel int) (*tgbotapi.MessageConfig, error) {
	updateRes := telegramUserModel.GetTelegramUserCollection().FindOneAndUpdate(
		context.Background(),
		bson.M{
			"telegramUserID": telegramUserID,
		},
		bson.M{
			"$set": bson.M{
				"accessLevel": accessLevel,
			},
		},
	)
	err := updateRes.Err()
	if err != nil {
		return nil, err
	}
	count, err := cache.DeleteUserSession(telegramUserID)
	fmt.Println(count, err)
	return telegramUtils.GenerateReplyMessage(update, "Updated "+strconv.Itoa(int(telegramUserID))+"'s Access Level: "+strconv.Itoa(int(accessLevel))+" ["+constants.AccessLevelToTextMap[accessLevel]+"]"), nil
}

func UpdateAccessLevel(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	argList := strings.Split(update.Message.CommandArguments(), " ")
	if len(argList) != 2 {
		return fmt.Errorf("invalid arguments")
	}

	telegramUserID, err := strconv.Atoi(argList[0])
	if err != nil {
		return err
	}

	accessLevel, err := strconv.Atoi(argList[1])
	if err != nil {
		return err
	}

	if accessLevel < constants.AccessLevelNone || accessLevel > constants.AccessLevelCreator {
		return fmt.Errorf("invalid access level")
	}

	requestedByTelegramUserID := update.Message.From.ID
	requestedByUserSession, err := telegramHelpers.GetUserSession(requestedByTelegramUserID)
	if err != nil {
		return err
	}

	// A minimum access level of Admin is required to update the access level
	requiredAccessLevel := utils.Max(accessLevel, constants.AccessLevelAdmin)

	if requestedByUserSession.AccessLevel < requiredAccessLevel {
		return telegramUtils.GenerateMinimumAccessLevelError(requestedByUserSession.AccessLevel, requiredAccessLevel)
	}

	msg, err := updateAccessLevel(update, int64(telegramUserID), accessLevel)
	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
