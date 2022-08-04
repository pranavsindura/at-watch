package telegramTypes

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Middleware func(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string) error
type ErrorMiddleware func(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string, err error)
type CommandMiddleware func(bot *tgbotapi.BotAPI, update tgbotapi.Update) error
