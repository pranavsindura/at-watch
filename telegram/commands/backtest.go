package telegramCommands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	backtestSDK "github.com/pranavsindura/at-watch/sdk/backtest"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	strategies "github.com/pranavsindura/at-watch/sdk/strategies/positional"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
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

func backtest(bot *telegramBot.BotAPI, update telegramBot.Update, strategy string, instrument string, timeFrame int, fromDate string, toDate string, lots int) {
	bot.Send(telegramUtils.GenerateReplyMessage(update, "Starting backtest"))
	backtestSDK := backtestSDK.NewBacktestSDK()

	switch strategy {
	case strategyConstants.StrategyPositional:
		totalProfit, totalLoss, totalBrokerage, profitTradeCount, lossTradeCount, totalTradeCount, finalPL, trades := strategies.BacktestPositionalStrategy(backtestSDK, instrument, timeFrame, fromDate, toDate, lots)
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
		tradesFileBytes := telegramBot.FileBytes{
			Name:  "trades.json",
			Bytes: []byte(tradesJSON),
		}
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
	if len(argList) != 6 {
		return fmt.Errorf("invalid arguments")
	}
	strategy := argList[0]
	instrument := argList[1]
	timeFrameText := argList[2]
	fromDate := argList[3]
	toDate := argList[4]
	lotsString := argList[5]

	fmt.Println(strategy, instrument, timeFrameText, fromDate, toDate, lotsString)

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
	if _, exists := marketConstants.TextToTimeFrameMap[timeFrameText]; !exists {
		return fmt.Errorf("invalid timeFrame")
	}
	if _, err := time.Parse(time.RFC3339, fromDate); fromDate == "" || err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339, toDate); toDate == "" || err != nil {
		return err
	}
	if _, err := strconv.Atoi(lotsString); err != nil {
		return fmt.Errorf("invalid lots")
	}

	timeFrame := marketConstants.TextToTimeFrameMap[timeFrameText]
	lots, _ := strconv.Atoi(lotsString)
	if lots < 0 {
		return fmt.Errorf("invalid lots")
	}

	backtest(bot, update, strategy, instrument, timeFrame, fromDate, toDate, lots)

	return nil
}
