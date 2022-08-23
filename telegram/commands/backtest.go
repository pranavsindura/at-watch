package telegramCommands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	backtestSDK "github.com/pranavsindura/at-watch/sdk/backtest"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	marketOpenStrategy "github.com/pranavsindura/at-watch/sdk/strategies/marketOpen"
	postionalStrategy "github.com/pranavsindura/at-watch/sdk/strategies/positional"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"github.com/yukithm/json2csv"
)

// func generateBacktestUsage() string {
// 	return "Invalid arguments, please use this format\n\n/" + telegramConstants.CommandBacktest + `
// 	<strategy>
// 	<instrument>
// 	<timeFrame>
// 	<fromDate>
// 	<toDate>
// 	<lotsPerTrade>`
// }

func backtest(bot *telegramBot.BotAPI, update telegramBot.Update, strategy string, instrument string, fromDate string, toDate string, lots int) {
	bot.Send(telegramUtils.GenerateReplyMessage(update, "Starting backtest"))
	backtestSDK := backtestSDK.NewBacktestSDK()

	switch strategy {
	case strategyConstants.StrategyPositional:
		totalProfit, totalLoss, totalBrokerage, profitTradeCount, lossTradeCount, totalTradeCount, finalPL, trades := postionalStrategy.BacktestPositionalStrategy(backtestSDK, instrument, fromDate, toDate, lots)
		text := `Results
		
totalProfit: ` + utils.RoundFloat(totalProfit) + `
totalLoss: ` + utils.RoundFloat(totalLoss) + `
totalBrokerage: ` + utils.RoundFloat(totalBrokerage) + `
profitTradeCount: ` + strconv.Itoa(int(profitTradeCount)) + `
lossTradeCount: ` + strconv.Itoa(int(lossTradeCount)) + `
totalTradeCount: ` + strconv.Itoa(int(totalTradeCount)) + `
finalPL: ` + utils.RoundFloat(finalPL)
		bot.Send(telegramUtils.GenerateReplyMessage(update, text))

		tradesJSON := utils.BruteStringify(trades)
		var tradesMapList []map[string]interface{}
		json.Unmarshal([]byte(tradesJSON), &tradesMapList)

		buff := &bytes.Buffer{}
		csvWriter := json2csv.NewCSVWriter(buff)
		csvWriter.HeaderStyle = json2csv.DotNotationStyle

		rows := make([]json2csv.KeyValue, 0)
		for _, tradesMap := range tradesMapList {
			row, err := json2csv.JSON2CSV(tradesMap)
			if err != nil {
				fmt.Println("attempting to convert to csv failed", row, err)
			}
			if len(row) != 1 {
				fmt.Println("unexpected row length", row)
				continue
			}
			rows = append(rows, row[0])
		}

		csvWriter.WriteCSV(rows)

		tradesFileBytes := telegramBot.FileBytes{
			Name:  "trades.csv",
			Bytes: buff.Bytes(),
		}
		docMsg := telegramUtils.GenerateReplyDocument(update, tradesFileBytes)
		bot.Send(docMsg)
	case strategyConstants.StrategyMarketOpen:
		totalProfit, totalLoss, totalBrokerage, profitTradeCount, lossTradeCount, totalTradeCount, finalPL, trades := marketOpenStrategy.BacktestMarketOpenStrategy(backtestSDK, instrument, fromDate, toDate, lots)
		text := `Results
		
totalProfit: ` + utils.RoundFloat(totalProfit) + `
totalLoss: ` + utils.RoundFloat(totalLoss) + `
totalBrokerage: ` + utils.RoundFloat(totalBrokerage) + `
profitTradeCount: ` + strconv.Itoa(int(profitTradeCount)) + `
lossTradeCount: ` + strconv.Itoa(int(lossTradeCount)) + `
totalTradeCount: ` + strconv.Itoa(int(totalTradeCount)) + `
finalPL: ` + utils.RoundFloat(finalPL)
		bot.Send(telegramUtils.GenerateReplyMessage(update, text))

		tradesJSON := utils.BruteStringify(trades)
		var tradesMapList []map[string]interface{}
		json.Unmarshal([]byte(tradesJSON), &tradesMapList)

		buff := &bytes.Buffer{}
		csvWriter := json2csv.NewCSVWriter(buff)
		csvWriter.HeaderStyle = json2csv.DotNotationStyle

		rows := make([]json2csv.KeyValue, 0)
		for _, tradesMap := range tradesMapList {
			row, err := json2csv.JSON2CSV(tradesMap)
			if err != nil {
				fmt.Println("attempting to convert to csv failed", row, err)
			}
			if len(row) != 1 {
				fmt.Println("unexpected row length", row)
				continue
			}
			rows = append(rows, row[0])
		}

		csvWriter.WriteCSV(rows)

		tradesFileBytes := telegramBot.FileBytes{
			Name:  "trades.csv",
			Bytes: buff.Bytes(),
		}
		// tradesFileBytes := telegramBot.FileBytes{
		// 	Name:  "trades.json",
		// 	Bytes: []byte(tradesJSON),
		// }
		docMsg := telegramUtils.GenerateReplyDocument(update, tradesFileBytes)
		bot.Send(docMsg)
	default:
		bot.Send(telegramUtils.GenerateReplyMessage(update, telegramUtils.GenerateGenericErrorText(fmt.Errorf("backtesting for this strategy does not exist"))))
	}
}

func Backtest(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	argString := update.Message.CommandArguments()
	argList := strings.Split(argString, "\n")
	fmt.Println(argList)
	if len(argList) != 5 {
		return fmt.Errorf("invalid arguments")
	}
	strategy := argList[0]
	instrument := argList[1]
	// timeFrameText := argList[2]
	fromDate := argList[2]
	toDate := argList[3]
	lotsString := argList[4]

	fmt.Println(strategy, instrument, fromDate, toDate, lotsString)

	isInstrumentValid, err := fyersSDK.IsValidInstrument(instrument)
	if err != nil {
		return err
	}

	if _, exists := strategyConstants.Strategies[strategy]; strategy == "" || !exists {
		return fmt.Errorf("invalid strategy")
	}
	if !isInstrumentValid {
		return fmt.Errorf("invalid instrument")
	}
	// if _, exists := marketConstants.TextToTimeFrameMap[timeFrameText]; !exists {
	// 	return fmt.Errorf("invalid timeFrame")
	// }
	if _, err := time.Parse(time.RFC3339, fromDate); fromDate == "" || err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339, toDate); toDate == "" || err != nil {
		return err
	}
	if _, err := strconv.Atoi(lotsString); err != nil {
		return fmt.Errorf("invalid lots")
	}

	// timeFrame := marketConstants.TextToTimeFrameMap[timeFrameText]
	lots, _ := strconv.Atoi(lotsString)
	if lots < 0 {
		return fmt.Errorf("invalid lots")
	}

	backtest(bot, update, strategy, instrument, fromDate, toDate, lots)

	return nil
}
