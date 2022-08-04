package telegramCommands

// import (
// 	"context"
// 	"fmt"

// 	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
// 	cache "github.com/pranavsindura/at-watch/cache"
// 	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
// 	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
// 	"go.mongodb.org/mongo-driver/bson"
// )

// func reset(update telegramBot.Update) telegramBot.MessageConfig {
// 	userName := update.Message.From.UserName

// 	_, err := cache.DeleteUserSession(userName)
// 	if err != nil {
// 		fmt.Println("unable to reset user", err)
// 		return *telegramUtils.GenerateReplyMessage(update, telegramUtils.GenerateGenericErrorText(err))
// 	}

// 	telegramUser := telegramUserModel.GetTelegramUserCollection()
// 	res, err := telegramUser.DeleteOne(context.Background(), bson.M{
// 		"userName": userName,
// 	})
// 	text := ""
// 	if err != nil {
// 		fmt.Println("unable to reset user", err)
// 		text = telegramUtils.GenerateGenericErrorText(err)
// 	} else if res.DeletedCount == 0 {
// 		text = "I dont know you"
// 	} else {
// 		text = "Done"
// 	}

// 	msg := telegramBot.NewMessage(update.Message.Chat.ID, text)
// 	msg.ReplyToMessageID = update.Message.MessageID
// 	return msg
// }

// func Reset(bot *telegramBot.BotAPI, update telegramBot.Update) {
// 	msg := reset(update)
// 	bot.Send(msg)
// }
