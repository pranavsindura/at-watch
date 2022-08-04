package telegramCommands

import (
	"context"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/constants"
	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/mongo"
)

func start(update telegramBot.Update) (*telegramBot.MessageConfig, error) {
	telegramUser := telegramUserModel.GetTelegramUserCollection()
	newUser := telegramUserModel.TelegramUserModel{
		FirstName:      update.Message.From.FirstName,
		LastName:       update.Message.From.LastName,
		TelegramUserID: update.Message.From.ID,
		TelegramChatID: update.Message.Chat.ID,
		AccessLevel:    constants.AccessLevelNewUser,
	}
	_, err := telegramUser.InsertOne(context.Background(), newUser)
	text := ""
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			text = "Welcome back to AlgoTrading, " + update.Message.From.FirstName + "!"
		} else {
			return nil, err
		}
	} else {
		text = "Welcome to AlgoTrading, " + update.Message.From.FirstName + "!"
	}

	msg := telegramUtils.GenerateReplyMessage(update, text)
	return msg, nil
}

func Start(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	msg, err := start(update)
	if err != nil {
		return err
	}
	bot.Send(msg)
	return nil
}
