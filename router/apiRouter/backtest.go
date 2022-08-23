package apiRouter

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	backtestSDK "github.com/pranavsindura/at-watch/sdk/backtest"
	positionalStrategy "github.com/pranavsindura/at-watch/sdk/strategies/positional"
	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	routerUtils "github.com/pranavsindura/at-watch/utils/router"
)

func backtest(instrument string, timeFrame int, fromDate string, toDate string, lots int) (gin.H, error) {
	backtestSDK := backtestSDK.NewBacktestSDK()
	pos := positionalStrategy.NewPositionalStrategy(instrument)
	backtestSDK.SubscribeCandle(instrument, marketConstants.TimeFrame15m, pos.CreateOnCandleForBacktest(func(candle *fyersTypes.FyersHistoricalCandle) int {
		return lots
	}))

	start, _ := time.Parse(time.RFC3339, fromDate)
	end, _ := time.Parse(time.RFC3339, toDate)
	backtestSDK.Backtest(instrument, timeFrame, start, end)

	loss := 0.
	lossCount := 0.
	profit := 0.
	profitCount := 0.
	brokerage := 0.
	finalPL := 0.

	for _, trade := range pos.ClosedTrades {
		if trade.PL > 0 {
			profitCount++
			profit += trade.PL
		} else if trade.PL < 0 {
			lossCount++
			loss += trade.PL
		}

		brokerage += trade.Brokerage

		finalPL += trade.PL - trade.Brokerage
	}

	fmt.Println("Total Profit", profit)
	fmt.Println("Total Loss", loss)
	fmt.Println("Total Brokerage", brokerage)
	fmt.Println("Profit making trades", profitCount, "/", profitCount+lossCount)
	fmt.Println("Final PL", finalPL)

	return gin.H{
		"totalProfit":      profit,
		"totalLoss":        loss,
		"totalBrokerage":   brokerage,
		"profitTradeCount": profitCount,
		"lossTradeCount":   lossCount,
		"totalTradeCount":  profitCount + lossCount,
		"finalPL":          finalPL,
		"trades":           pos.ClosedTrades,
	}, nil
}

func Backtest(ctx *gin.Context) {
	instrument, _ := ctx.GetQuery("instrument")
	timeFrameText, _ := ctx.GetQuery("timeFrame")
	fromDate, _ := ctx.GetQuery("fromDate")
	toDate, _ := ctx.GetQuery("toDate")
	lotsString, _ := ctx.GetQuery("lots")

	// if _, exists := marketConstants.Instruments[instrument]; instrument == "" || !exists {
	// 	routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid instrument"))
	// 	return
	// }
	if _, exists := marketConstants.TextToTimeFrameMap[timeFrameText]; !exists {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid timeFrame"))
		return
	}
	if _, err := time.Parse(time.RFC3339, fromDate); fromDate == "" || err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}
	if _, err := time.Parse(time.RFC3339, toDate); toDate == "" || err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid toDate"))
		return
	}
	if _, err := strconv.Atoi(lotsString); err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid lots"))
		return
	}

	timeFrame := marketConstants.TextToTimeFrameMap[timeFrameText]
	lots, _ := strconv.Atoi(lotsString)

	data, err := backtest(instrument, timeFrame, fromDate, toDate, lots)

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	routerUtils.SendSuccessResponse(ctx, http.StatusOK, data)
}
