package backtestSDK

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	marketTypes "github.com/pranavsindura/at-watch/types/market"
	"github.com/pranavsindura/at-watch/utils"
	marketUtils "github.com/pranavsindura/at-watch/utils/market"
	eventemitter "github.com/vansante/go-event-emitter"
)

type BacktestSDK struct {
	backtestEventEmitter *eventemitter.Emitter
	partialCandle15m     map[string]*fyersTypes.FyersHistoricalCandle
}

func NewBacktestSDK() *BacktestSDK {
	sdk := &BacktestSDK{
		backtestEventEmitter: eventemitter.NewEmitter(true),
		partialCandle15m:     make(map[string]*fyersTypes.FyersHistoricalCandle),
	}
	return sdk
}

// var backtestEventEmitter = eventemitter.NewEmitter(true)
// var partialCandle15m map[string]*fyersTypes.FyersHistoricalCandle = make(map[string]*fyersTypes.FyersHistoricalCandle)

func (sdk *BacktestSDK) generateEventKey(event string, instrument string, timeFrame int) eventemitter.EventType {
	return eventemitter.EventType(event + ":" + instrument + ":" + strconv.Itoa(timeFrame))
}

func (sdk *BacktestSDK) SubscribeTick(instrument string, timeFrame int, callback func(fyersTypes.FyersHistoricalCandle)) func() {
	key := sdk.generateEventKey(marketConstants.MarketEventTick, instrument, timeFrame)

	var handlerFn eventemitter.HandleFunc = func(arguments ...interface{}) {
		var candle fyersTypes.FyersHistoricalCandle = arguments[0].(fyersTypes.FyersHistoricalCandle)
		callback(candle)
	}

	listener := sdk.backtestEventEmitter.AddListener(key, handlerFn)

	unsubscribeFn := func() {
		fmt.Println("stopping tick", key)
		sdk.backtestEventEmitter.RemoveListener(key, listener)
	}

	return unsubscribeFn
}

func (sdk *BacktestSDK) emitTick(instrument string, timeFrame int, candle fyersTypes.FyersHistoricalCandle) {
	key := sdk.generateEventKey(marketConstants.MarketEventTick, instrument, timeFrame)
	copyCandle := fyersTypes.FyersHistoricalCandle{}
	err := copier.Copy(&copyCandle, &candle)

	if err != nil {
		fmt.Println("Not able to copy candle for tick:", candle)
	}

	sdk.backtestEventEmitter.EmitEvent(key, copyCandle)
}

func (sdk *BacktestSDK) SubscribeCandle(instrument string, timeFrame int, callback func(*fyersTypes.FyersHistoricalCandle)) func() {
	key := sdk.generateEventKey(marketConstants.MarketEventCandle, instrument, timeFrame)

	var handlerFn eventemitter.HandleFunc = func(arguments ...interface{}) {
		var candle fyersTypes.FyersHistoricalCandle = arguments[0].(fyersTypes.FyersHistoricalCandle)
		callback(&candle)
	}

	listener := sdk.backtestEventEmitter.AddListener(key, handlerFn)

	unsubscribeFn := func() {
		fmt.Println("stopping candle", key)
		sdk.backtestEventEmitter.RemoveListener(key, listener)
	}

	return unsubscribeFn
}

func (sdk *BacktestSDK) emitCandle(instrument string, timeFrame int, candle fyersTypes.FyersHistoricalCandle) {
	key := sdk.generateEventKey(marketConstants.MarketEventCandle, instrument, timeFrame)
	copyCandle := fyersTypes.FyersHistoricalCandle{}
	err := copier.Copy(&copyCandle, &candle)

	if err != nil {
		fmt.Println("Not able to copy candle for candle:", candle)
	}

	sdk.backtestEventEmitter.EmitEvent(key, copyCandle)
}

func (sdk *BacktestSDK) getPartialCandle(instrument string, timeFrame int) *fyersTypes.FyersHistoricalCandle {
	switch timeFrame {
	case marketConstants.TimeFrame15m:
		candle, exists := sdk.partialCandle15m[instrument]
		if !exists {
			return nil
		}
		var copyCandle *fyersTypes.FyersHistoricalCandle = &fyersTypes.FyersHistoricalCandle{}
		copier.Copy(copyCandle, candle)
		return copyCandle
	}
	return nil
}

func (sdk *BacktestSDK) setPartialCandle(instrument string, timeFrame int, candle *fyersTypes.FyersHistoricalCandle) {
	switch timeFrame {
	case marketConstants.TimeFrame15m:
		var copyCandle *fyersTypes.FyersHistoricalCandle = &fyersTypes.FyersHistoricalCandle{}
		copier.Copy(copyCandle, candle)
		delete(sdk.partialCandle15m, instrument)
		sdk.partialCandle15m[instrument] = candle
	}
}

func (sdk *BacktestSDK) updateCandle(instrument string, timeFrame int, tickData marketTypes.MarketTick) (*fyersTypes.FyersHistoricalCandle, *fyersTypes.FyersHistoricalCandle) {
	var lastCandle *fyersTypes.FyersHistoricalCandle = sdk.getPartialCandle(instrument, timeFrame)
	var currentTickCandleTimestamp time.Time = marketUtils.GetCandleTimeOf(timeFrame, time.Unix(tickData.TS, 0))

	var partialCandle *fyersTypes.FyersHistoricalCandle = nil
	var completeCandle *fyersTypes.FyersHistoricalCandle = nil

	ltp := tickData.LTP
	volume := tickData.Volume

	if lastCandle != nil {
		if lastCandle.TS != currentTickCandleTimestamp.Unix() {
			// new candle, ship off the last candle, make a new one
			newCandle := fyersTypes.FyersHistoricalCandle{
				TS:         currentTickCandleTimestamp.Unix(),
				Open:       ltp,
				Close:      ltp,
				High:       ltp,
				Low:        ltp,
				DateString: utils.GetDateStringFromTimestamp(currentTickCandleTimestamp.Unix()),
				Day:        utils.GetWeekdayFromTimestamp(currentTickCandleTimestamp.Unix()),
				Volume:     volume,
			}
			partialCandle = &fyersTypes.FyersHistoricalCandle{}
			completeCandle = &fyersTypes.FyersHistoricalCandle{}
			copier.Copy(completeCandle, lastCandle)
			copier.Copy(partialCandle, newCandle)
			sdk.setPartialCandle(instrument, timeFrame, &newCandle)
		} else {
			// update same candle
			lastCandle.Close = ltp
			lastCandle.High = math.Max(lastCandle.High, ltp)
			lastCandle.Low = math.Min(lastCandle.Low, ltp)
			partialCandle = &fyersTypes.FyersHistoricalCandle{}
			copier.Copy(partialCandle, lastCandle)
			sdk.setPartialCandle(instrument, timeFrame, lastCandle)
		}
	} else {
		// first candle
		firstCandle := fyersTypes.FyersHistoricalCandle{
			TS:         currentTickCandleTimestamp.Unix(),
			Open:       ltp,
			Close:      ltp,
			High:       ltp,
			Low:        ltp,
			DateString: utils.GetDateStringFromTimestamp(currentTickCandleTimestamp.Unix()),
			Day:        utils.GetWeekdayFromTimestamp(currentTickCandleTimestamp.Unix()),
			Volume:     volume,
		}
		partialCandle = &fyersTypes.FyersHistoricalCandle{}
		copier.Copy(partialCandle, firstCandle)
		sdk.setPartialCandle(instrument, timeFrame, &firstCandle)
	}

	partialCandle.TS = tickData.TS
	partialCandle.DateString = utils.GetDateStringFromTimestamp(tickData.TS)
	partialCandle.Day = utils.GetWeekdayFromTimestamp(tickData.TS)

	return partialCandle, completeCandle
}

func (sdk *BacktestSDK) Backtest(instrument string, timeFrame int, fromDate time.Time, toDate time.Time) {
	const resolution string = marketConstants.Resolution1m

	currentDate := time.Time{}
	copier.Copy(&currentDate, &fromDate)

	var allCandles [](fyersTypes.FyersHistoricalCandle) = make([]fyersTypes.FyersHistoricalCandle, 0)

	fmt.Println(fromDate, toDate)

	for currentDate.Sub(toDate).Milliseconds() <= 0 {
		after3Months := currentDate.AddDate(0, 3, 0)
		if toDate.Sub(after3Months).Milliseconds() < 0 {
			after3Months = toDate
		}

		fmt.Println("fetching", utils.GetDateStringFromTimestamp(currentDate.Unix()), "to", utils.GetDateStringFromTimestamp(after3Months.Unix()))

		data, err := fyersSDK.FetchHistoricalData(instrument, resolution, currentDate.Unix(), after3Months.Unix(), 0)
		if err != nil {
			fmt.Println("Error while fetching historical data", err)
		}

		allCandles = append(allCandles, data...)

		fmt.Println("got", instrument, len(data))

		currentDate = after3Months.AddDate(0, 0, 1)
	}

	fmt.Println("got total", instrument, len(allCandles))

	/**
	 * tick should be sent only after the candles are updated, this is important
	 * but ticks are sent before candle events
	 */
	for _, tick := range allCandles {
		partialCandle1, completeCandle1 := sdk.updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.Open, Volume: tick.Volume})
		sdk.emitTick(instrument, timeFrame, *partialCandle1)
		if completeCandle1 != nil {
			sdk.emitCandle(instrument, timeFrame, *completeCandle1)
		}

		partialCandle2, completeCandle2 := sdk.updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.High, Volume: tick.Volume})
		sdk.emitTick(instrument, timeFrame, *partialCandle2)
		if completeCandle2 != nil {
			sdk.emitCandle(instrument, timeFrame, *completeCandle2)
		}

		partialCandle3, completeCandle3 := sdk.updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.Low, Volume: tick.Volume})
		sdk.emitTick(instrument, timeFrame, *partialCandle3)
		if completeCandle3 != nil {
			sdk.emitCandle(instrument, timeFrame, *completeCandle3)
		}

		partialCandle4, completeCandle4 := sdk.updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.Close, Volume: tick.Volume})
		sdk.emitTick(instrument, timeFrame, *partialCandle4)
		if completeCandle4 != nil {
			sdk.emitCandle(instrument, timeFrame, *completeCandle4)
		}
	}
}
