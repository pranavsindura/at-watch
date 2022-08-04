package telegramCommands

import (
	"context"
	"fmt"
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

func getClosedTrade(update tgbotapi.Update, userID primitive.ObjectID) (*tgbotapi.MessageConfig, error) {
	text := ""
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
		return nil, err
	}

	isFirstPosClosedTrade := true
	for posClosedTradesCursor.Next(context.Background()) {
		if isFirstPosClosedTrade {
			text += strategyConstants.StrategyPositional + "\n"
			isFirstPosClosedTrade = false
		}
		closedTradeCount++
		aggResult := PositionalStrategyAggregateResult{}
		posClosedTradesCursor.Decode(&aggResult)
		text += "Strategy ID: " + aggResult.ID.Hex() + "\n"
		text += "Instrument: " + aggResult.Instrument.Symbol + "\n"
		text += "Trade Type: " + aggResult.Trade.TradeTypeText + "\n"
		text += "Lots: " + strconv.Itoa(aggResult.Trade.Lots) + "\n"
		text += "Entry At: " + aggResult.Trade.Entry.Candle.DateString + "\n"
		text += "Exit At: " + aggResult.Trade.Exit.Candle.DateString + "\n"
		text += "Entry Price: " + fmt.Sprintf("%.2f", aggResult.Trade.Entry.Candle.Close) + "\n"
		text += "Exit Price: " + fmt.Sprintf("%.2f", aggResult.Trade.Exit.Candle.Close) + "\n"
		text += "PL: " + utils.RoundFloat(aggResult.Trade.PL) + "\n"
		text += "Brokerage: " + fmt.Sprintf("%.2f", aggResult.Trade.Brokerage) + "\n"
		text += "Exit Reason: " + aggResult.Trade.ExitReasonText + "\n"
		text += "\n"
	}

	text = strconv.Itoa(closedTradeCount) + " Closed Trades\n\n" + text

	return telegramUtils.GenerateReplyMessage(update, text), nil
}

func GetClosedTrade(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	telegramUserID := update.Message.From.ID
	userSession, err := telegramHelpers.GetUserSession(telegramUserID)
	if err != nil {
		return err
	}

	userID := userSession.UserID

	msg, err := getClosedTrade(update, userID)

	if err != nil {
		return err
	}

	bot.Send(msg)
	return nil
}
