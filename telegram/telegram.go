package telegram

import (
	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	telegramClient "github.com/pranavsindura/at-watch/connections/telegram"
	telegramConstants "github.com/pranavsindura/at-watch/constants/telegram"
	telegramCommands "github.com/pranavsindura/at-watch/telegram/commands"

	telegramMiddlewares "github.com/pranavsindura/at-watch/telegram/middlewares"
	telegramTypes "github.com/pranavsindura/at-watch/types/telegram"
)

var DefaultErrorHandler = telegramMiddlewares.ErrorHandler

func Execute(command string, bot *telegramBot.BotAPI, update telegramBot.Update, middlewares []telegramTypes.Middleware, commandMiddleware telegramTypes.CommandMiddleware) {
	for _, middleware := range middlewares {
		err := middleware(bot, update, command)
		if err != nil {
			DefaultErrorHandler(bot, update, command, err)
			return
		}
	}
	err := commandMiddleware(bot, update)
	if err != nil {
		DefaultErrorHandler(bot, update, command, err)
	}
}

func HandleUpdate(bot *telegramBot.BotAPI, update telegramBot.Update) {
	if update.Message.From.IsBot || !update.Message.Chat.IsPrivate() {
		return
	}

	execute := func(middlewares []telegramTypes.Middleware, commandMiddleware telegramTypes.CommandMiddleware) {
		Execute(update.Message.Command(), bot, update, middlewares, commandMiddleware)
	}

	command := update.Message.Command()

	switch command {
	case telegramConstants.CommandPing:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.Ping)

	case telegramConstants.CommandStart:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.Start)

	case telegramConstants.CommandStop:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.Stop)

	case telegramConstants.CommandLogin:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.Login)

	case telegramConstants.CommandAdminLogin:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotActiveAndNotWarmingUp}, telegramCommands.AdminLogin)

	case telegramConstants.CommandMe:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.Me)

	case telegramConstants.CommandUpdateAccessLevel:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.UpdateAccessLevel)

	case telegramConstants.CommandBacktest:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.FyersAccessTokenExists}, telegramCommands.Backtest)

	case telegramConstants.CommandMaintenance:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.Maintenance)

	case telegramConstants.CommandAddInstrument:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotActiveAndNotWarmingUp}, telegramCommands.AddInstrument)

	case telegramConstants.CommandRemoveInstrument:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotActiveAndNotWarmingUp}, telegramCommands.RemoveInstrument)

	case telegramConstants.CommandGetInstrument:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.GetInstrument)

	case telegramConstants.CommandRenameInstrument:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.RenameInstrument)

	case telegramConstants.CommandAddStrategy:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotActiveAndNotWarmingUp}, telegramCommands.AddStrategy)

	case telegramConstants.CommandGetStrategy:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.GetStrategy)

	case telegramConstants.CommandRemoveStrategy:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotActiveAndNotWarmingUp}, telegramCommands.RemoveStrategy)

	case telegramConstants.CommandPauseStrategy:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotActiveAndNotWarmingUp}, telegramCommands.PauseStrategy)

	case telegramConstants.CommandResumeStrategy:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotActiveAndNotWarmingUp}, telegramCommands.ResumeStrategy)

	case telegramConstants.CommandStartMarket:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.StartMarket)

	case telegramConstants.CommandStopMarket:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.StopMarket)

	case telegramConstants.CommandEnterStrategy:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.EnterStrategy)

	case telegramConstants.CommandExitStrategy:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.ExitStrategy)

	case telegramConstants.CommandGetClosedTrade:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler}, telegramCommands.GetClosedTrade)

	case telegramConstants.CommandGetOpenTrade:
		execute([]telegramTypes.Middleware{telegramMiddlewares.AccessLevelHandler, telegramMiddlewares.MarketNotWarmingUp}, telegramCommands.GetOpenTrade)
	}
}

func FetchUpdatesAndHandle() {
	u := telegramBot.NewUpdate(0)
	u.Timeout = 60

	updates := telegramClient.Client().GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message != nil {
				go HandleUpdate(telegramClient.Client(), update)
			}
		}
	}()
}
