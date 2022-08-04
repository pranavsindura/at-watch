package telegramCommands

import (
	"context"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/constants"
	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
)

func me(update tgbotapi.Update, telegramUserID int64) (*tgbotapi.MessageConfig, error) {
	res := telegramUserModel.GetTelegramUserCollection().FindOne(context.Background(), bson.M{
		"telegramUserID": telegramUserID,
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}
	user := telegramUserModel.TelegramUserModel{}
	res.Decode(&user)

	text := "Name: " + user.FirstName + " " + user.LastName + "\n"
	text += "TelegramUserID: " + strconv.Itoa(int(user.TelegramUserID)) + "\n"
	text += "TelegramChatID: " + strconv.Itoa(int(user.TelegramChatID)) + "\n"
	text += "AccessLevel: " + strconv.Itoa(int(user.AccessLevel)) + "\n"
	text += "AccessLevelText: " + constants.AccessLevelToTextMap[user.AccessLevel] + "\n"
	text += "UserID: " + user.ID.Hex() + "\n"

	return telegramUtils.GenerateReplyMessage(update, text), nil
}

func Me(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	telegramUserID := update.Message.From.ID

	msg, err := me(update, telegramUserID)
	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
