package telegramCommands

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
)

// func generateGetInstrumentUsage() string {
// 	return "Invalid arguments, please use this format\n\n/" + telegramConstants.CommandGetInstrument
// }

func getInstrument(update tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	instruments, err := telegramHelpers.GetInstruments()

	fmt.Println(instruments, len(instruments))

	if err != nil {
		return nil, err
	}

	text := strings.Join(instruments, "\n")

	return telegramUtils.GenerateReplyMessage(update, strconv.Itoa(len(instruments))+" Instruments\n\n"+text), nil
}

func GetInstrument(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	msg, err := getInstrument(update)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
