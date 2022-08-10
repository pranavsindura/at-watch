package telegramCommands

import (
	"context"
	"fmt"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cache "github.com/pranavsindura/at-watch/cache"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	positionalTradeModel "github.com/pranavsindura/at-watch/models/positionalTrade"
	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
)

func stop(update telegramBot.Update) (*telegramBot.MessageConfig, error) {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		fmt.Println("error while fetching user session")
		return nil, err
	}

	// Remove strategies from marketSDK
	// Remove all Open/Closed Trades
	// Remove all Strategies
	// Remove user data

	// Positional
	positionalStrategiesCursor, err := positionalStrategyModel.
		GetPositionalStrategyCollection().
		Find(context.Background(), bson.M{
			"userID": userSession.UserID,
		})
	if err != nil {
		fmt.Println("error while fetching positional strategies")
		return nil, err
	}
	positionalStrategies := make([]positionalStrategyModel.PositionalStrategy, 0)
	positionalStrategiesCursor.All(context.Background(), &positionalStrategies)

	for _, strategy := range positionalStrategies {
		if strategy.IsActive {
			marketSDK.RemovePositionalStrategy(strategy.ID)
		}
		_, err = positionalTradeModel.
			GetPositionalTradeCollection().
			DeleteMany(context.Background(), bson.M{
				"strategyID": strategy.ID,
			})
		if err != nil {
			fmt.Println("error while deleting positional trades for strategy id - " + strategy.ID.Hex())
			return nil, err
		}
	}

	_, err = positionalStrategyModel.
		GetPositionalStrategyCollection().
		DeleteMany(context.Background(), bson.M{
			"userID": userSession.UserID,
		})
	if err != nil {
		fmt.Println("error while deleting positional strategies for user id - " + userSession.UserID.Hex())
		return nil, err
	}

	_, err = telegramUserModel.GetTelegramUserCollection().DeleteOne(context.Background(), bson.M{"telegramUserID": telegramUserID})
	if err != nil {
		fmt.Println("error while deleting user - " + userSession.UserID.Hex())
		return nil, err
	}
	// Clear cache
	cache.DeleteUserSession(telegramUserID)
	return telegramUtils.GenerateReplyMessage(update, "Successfully deleted all data"), nil
}

func Stop(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	msg, err := stop(update)
	if err != nil {
		return err
	}
	bot.Send(msg)
	return nil
}
