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
	"github.com/pranavsindura/at-watch/sdk/notifications"
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
	updatedUser := telegramUserModel.TelegramUserModel{}
	updateRes.Decode(&updatedUser)
	cache.DeleteUserSession(telegramUserID)

	notifyText := "Your Updated Access Level: " + strconv.Itoa(int(accessLevel)) + " [" + constants.AccessLevelToTextMap[accessLevel] + "]\n"
	msgText := "Updated " + strconv.Itoa(int(telegramUserID)) + "'s Access Level: " + strconv.Itoa(int(accessLevel)) + " [" + constants.AccessLevelToTextMap[accessLevel] + "]\n"

	// if accessLevel was == constants.AccessLevelNewUser
	// execute the /stop flow
	if accessLevel == constants.AccessLevelNewUser {
		_, err := stop(update)
		if err != nil {
			fmt.Println("error while performing /stop, access level was " + constants.AccessLevelNewUserText)
			return nil, err
		}
		notifyText += "Successfully stop all data\n"
		msgText += "Successfully stop all data\n"
	}

	notifications.Notify(updatedUser.TelegramChatID, notifyText)
	return telegramUtils.GenerateReplyMessage(update, msgText), nil
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

	if accessLevel < constants.AccessLevelNewUser || accessLevel > constants.AccessLevelCreator {
		return fmt.Errorf("invalid access level")
	}

	requestedByTelegramUserID := update.Message.From.ID
	requestedByUserSession, err := telegramHelpers.GetUserSession(requestedByTelegramUserID)
	if err != nil {
		return err
	}

	// A minimum access level of Admin is required to update the access level
	// or requested+1
	requiredAccessLevel := utils.Max(accessLevel+1, constants.AccessLevelAdmin)

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
