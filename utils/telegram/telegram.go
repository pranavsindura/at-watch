package telegramUtils

import (
	"fmt"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pranavsindura/at-watch/constants"
)

func GenerateGenericErrorText(err error) string {
	return "unable to perform the action\n\n" + err.Error() + "\n\nplease contact @pranavsindura if the issue persists"
}

func GenerateMinimumAccessLevelError(currentAccessLevel, requiredAccessLevel int) error {
	return fmt.Errorf("action requires access level " + constants.AccessLevelToTextMap[requiredAccessLevel] + ", but you have access level " + constants.AccessLevelToTextMap[currentAccessLevel])
}

func GenerateReplyMessage(update telegramBot.Update, text string) *telegramBot.MessageConfig {
	msg := telegramBot.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID
	return &msg
}

func GenerateReplyDocument(update telegramBot.Update, doc telegramBot.RequestFileData) *telegramBot.DocumentConfig {
	msg := telegramBot.NewDocument(update.Message.Chat.ID, doc)
	msg.ReplyToMessageID = update.Message.MessageID
	return &msg
}
