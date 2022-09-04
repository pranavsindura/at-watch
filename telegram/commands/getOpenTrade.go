package telegramCommands

import (
	"context"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	mongoConstants "github.com/pranavsindura/at-watch/constants/mongo"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	instrumentModel "github.com/pranavsindura/at-watch/models/instrument"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	positionalTradeModel "github.com/pranavsindura/at-watch/models/positionalTrade"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getOpenTrade(update tgbotapi.Update, userID primitive.ObjectID) (*tgbotapi.MessageConfig, error) {
	text := ""
	openTradeCount := 0
	if marketSDK.IsMarketWatchActive() {
		// use tick based ltp

		// Positional
		posOpenTrades, posStrategies, err := marketSDK.GetPositionalOpenTrades(userID)
		if err != nil {
			return nil, err
		}

		openTradeCount += len(posOpenTrades)
		if len(posOpenTrades) > 0 {
			text += strategyConstants.StrategyPositional + "\n"
		}

		for idx, openTrade := range posOpenTrades {
			text += "Strategy ID: " + posStrategies[idx].ID.Hex() + "\n"
			text += "Instrument: " + posStrategies[idx].Instrument + "\n"
			text += "Trade Type: " + openTrade.TradeTypeText + "\n"
			text += "Entry At: " + openTrade.Entry.Candle.DateString + "\n"
			text += "Entry Price: " + utils.RoundFloat(openTrade.Entry.Candle.Close) + "\n"
			text += "Lots: " + strconv.Itoa(openTrade.Lots) + "\n"
			text += "PL: " + utils.RoundFloat(openTrade.PL) + "\n"
			text += "LTP: " + utils.RoundFloat(openTrade.UpdatedAtLTP) + "\n"
			text += "Updated At: " + utils.GetDateStringFromTimestamp(openTrade.UpdatedAtTS) + "\n"
			text += "\n"
		}
	} else {
		// Positional
		type PositionalStrategyAggregateResult struct {
			positionalStrategyModel.PositionalStrategy `bson:",inline"`

			Trade      positionalTradeModel.PositionalTrade `bson:"trade"`
			Instrument instrumentModel.InstrumentModel      `bson:"instrument"`
		}
		posMatchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "userID", Value: userID}}}}
		posLookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: mongoConstants.PositionalTradeCollection}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "strategyID"}, {Key: "as", Value: "trade"}}}}
		posUnwindStage := bson.D{{Key: "$unwind", Value: "$trade"}}
		posMatchStage2 := bson.D{{Key: "$match", Value: bson.D{{Key: "trade.status", Value: marketConstants.TradeStatusOpen}}}}
		posLookupStage2 := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: mongoConstants.InstrumentCollection}, {Key: "localField", Value: "instrumentID"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "instrument"}}}}
		posUnwindStage2 := bson.D{{Key: "$unwind", Value: "$instrument"}}
		posOpenTradesCursor, err := positionalStrategyModel.
			GetPositionalStrategyCollection().
			Aggregate(
				context.Background(),
				mongo.Pipeline{
					posMatchStage,
					posLookupStage,
					posUnwindStage,
					posMatchStage2,
					posLookupStage2,
					posUnwindStage2,
				},
			)
		if err != nil {
			return nil, err
		}
		for posOpenTradesCursor.Next(context.Background()) {
			openTradeCount++
			aggResult := PositionalStrategyAggregateResult{}
			posOpenTradesCursor.Decode(&aggResult)
			text += "Strategy ID: " + aggResult.ID.Hex() + "\n"
			text += "Instrument: " + aggResult.Instrument.Symbol + "\n"
			text += "Trade Type: " + aggResult.Trade.TradeTypeText + "\n"
			text += "Entry At: " + aggResult.Trade.Entry.Candle.DateString + "\n"
			text += "Entry Price: " + utils.RoundFloat(aggResult.Trade.Entry.Candle.Close) + "\n"
			text += "Lots: " + strconv.Itoa(aggResult.Trade.Lots) + "\n"
			text += "PL: " + utils.RoundFloat(aggResult.Trade.PL) + "\n"
			text += "Updated At: " + utils.GetDateStringFromTimestamp(aggResult.Trade.UpdatedAt) + "\n"
			text += "\n"
		}
	}

	text = strconv.Itoa(openTradeCount) + " Open Trades\n\n" + text

	return telegramUtils.GenerateReplyMessage(update, text), nil
}

func GetOpenTrade(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return err
	}

	userID := userSession.UserID

	msg, err := getOpenTrade(update, userID)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
