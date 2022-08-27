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
	telegramHelpers "github.com/pranavsindura/at-watch/telegram/helpers"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getClosedTrade(bot *tgbotapi.BotAPI, update tgbotapi.Update, userID primitive.ObjectID) error {
	closedTradeCount := 0
	// Positional
	type PositionalStrategyAggregateResult struct {
		positionalStrategyModel.PositionalStrategy `bson:",inline"`

		Trade      positionalTradeModel.PositionalTrade `bson:"trade"`
		Instrument instrumentModel.InstrumentModel      `bson:"instrument"`
	}
	posMatchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "userID", Value: userID}}}}
	posLookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: mongoConstants.PositionalTradeCollection}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "strategyID"}, {Key: "as", Value: "trade"}}}}
	posUnwindStage := bson.D{{Key: "$unwind", Value: "$trade"}}
	posMatchStage2 := bson.D{{Key: "$match", Value: bson.D{{Key: "trade.status", Value: marketConstants.TradeStatusClosed}}}}
	posLookupStage2 := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: mongoConstants.InstrumentCollection}, {Key: "localField", Value: "instrumentID"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "instrument"}}}}
	posUnwindStage2 := bson.D{{Key: "$unwind", Value: "$instrument"}}
	posClosedTradesCursor, err := positionalStrategyModel.
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
		return err
	}

	type PositionalStrategySendResult struct {
		StrategyID string `json:"strategyID"`
		Instrument string `json:"instrument"`
		TradeType  string `json:"tradeType"`
		Lots       string `json:"lots"`
		EntryAt    string `json:"entryAt"`
		ExitAt     string `json:"exitAt"`
		EntryPrice string `json:"entryPrice"`
		ExitPrice  string `json:"exitPrice"`
		PL         string `json:"PL"`
		Brokerage  string `json:"brokerage"`
		ExitReason string `json:"exitReason"`
	}
	var posClosedTrades []PositionalStrategySendResult = make([]PositionalStrategySendResult, 0)
	for posClosedTradesCursor.Next(context.Background()) {
		aggResult := PositionalStrategyAggregateResult{}
		posClosedTradesCursor.Decode(&aggResult)

		trade := PositionalStrategySendResult{
			StrategyID: aggResult.ID.Hex(),
			Instrument: aggResult.Instrument.Symbol,
			TradeType:  aggResult.Trade.TradeTypeText,
			Lots:       strconv.Itoa(aggResult.Trade.Lots),
			EntryAt:    aggResult.Trade.Entry.Candle.DateString,
			ExitAt:     aggResult.Trade.Exit.Candle.DateString,
			EntryPrice: utils.RoundFloat(aggResult.Trade.Entry.Candle.Close),
			ExitPrice:  utils.RoundFloat(aggResult.Trade.Exit.Candle.Close),
			PL:         utils.RoundFloat(aggResult.Trade.PL),
			Brokerage:  utils.RoundFloat(aggResult.Trade.Brokerage),
			ExitReason: aggResult.Trade.ExitReasonText,
		}
		posClosedTrades = append(posClosedTrades, trade)
	}

	if len(posClosedTrades) > 0 {
		closedTradeCount += len(posClosedTrades)
		posClosedTradesFile := tgbotapi.FileBytes{
			Name:  strategyConstants.StrategyPositional + ".csv",
			Bytes: utils.JSONList2CSVBytes(utils.BruteStringify(posClosedTrades)),
		}
		bot.Send(telegramUtils.GenerateReplyDocument(update, posClosedTradesFile))
	}

	bot.Send(telegramUtils.GenerateReplyMessage(update, strconv.Itoa(closedTradeCount)+" Closed Trades"))

	return nil
}

func GetClosedTrade(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return err
	}

	userID := userSession.UserID

	err = getClosedTrade(bot, update, userID)

	if err != nil {
		return err
	}

	return nil
}
