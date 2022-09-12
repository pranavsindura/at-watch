package marketSDK

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/pranavsindura/at-watch/constants"
	fyersConstants "github.com/pranavsindura/at-watch/constants/fyers"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	strategyConstants "github.com/pranavsindura/at-watch/constants/strategies"
	instrumentModel "github.com/pranavsindura/at-watch/models/instrument"
	positionalStrategyModel "github.com/pranavsindura/at-watch/models/positionalStrategy"
	positionalTradeModel "github.com/pranavsindura/at-watch/models/positionalTrade"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	fyersWatchAPI "github.com/pranavsindura/at-watch/sdk/fyersWatch/api"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	postionalStrategy "github.com/pranavsindura/at-watch/sdk/strategies/positional"
	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	indicatorTypes "github.com/pranavsindura/at-watch/types/indicators"
	marketTypes "github.com/pranavsindura/at-watch/types/market"
	"github.com/pranavsindura/at-watch/utils"
	marketUtils "github.com/pranavsindura/at-watch/utils/market"
	eventemitter "github.com/vansante/go-event-emitter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var isMarketActiveMutex *sync.RWMutex = &sync.RWMutex{}
var isMarketActive = false
var isWarmUpInProgressMutex *sync.RWMutex = &sync.RWMutex{}
var isWarmUpInProgress = false
var marketEventEmitter *eventemitter.Emitter = eventemitter.NewEmitter(true)
var partialCandle15m = make(map[string]*fyersTypes.FyersHistoricalCandle)
var partialCandle15mMutex = &sync.RWMutex{}

var strategyIDToPositionalStrategyMap = make(map[primitive.ObjectID]*postionalStrategy.PositionalStrategy)

var strategyIDToTickSubscription = make(map[primitive.ObjectID]func())
var strategyIDToCandleSubscription = make(map[primitive.ObjectID]func())

var fyersWatchNotificationChannel chan fyersWatchAPI.Notification = nil

func generateEventKey(event string, instrument string, timeFrame int) eventemitter.EventType {
	return eventemitter.EventType(event + ":" + instrument + ":" + strconv.Itoa(timeFrame))
}

func SubscribeTick(instrument string, timeFrame int, callback func(*fyersTypes.FyersHistoricalCandle)) func() {
	key := generateEventKey(marketConstants.MarketEventTick, instrument, timeFrame)

	var handlerFn eventemitter.HandleFunc = func(arguments ...interface{}) {
		var candle fyersTypes.FyersHistoricalCandle = arguments[0].(fyersTypes.FyersHistoricalCandle)
		callback(&candle)
	}

	listener := marketEventEmitter.AddListener(key, handlerFn)

	unsubscribeFn := func() {
		fmt.Println("stopping tick", key)
		marketEventEmitter.RemoveListener(key, listener)
	}

	return unsubscribeFn
}

func emitTick(instrument string, timeFrame int, candle fyersTypes.FyersHistoricalCandle) {
	key := generateEventKey(marketConstants.MarketEventTick, instrument, timeFrame)
	copyCandle := fyersTypes.FyersHistoricalCandle{}
	err := copier.Copy(&copyCandle, &candle)

	if err != nil {
		fmt.Println("Not able to copy candle for tick:", candle)
	}

	marketEventEmitter.EmitEvent(key, copyCandle)
}

func SubscribeCandle(instrument string, timeFrame int, callback func(candle *fyersTypes.FyersHistoricalCandle)) func() {
	key := generateEventKey(marketConstants.MarketEventCandle, instrument, timeFrame)

	var handlerFn eventemitter.HandleFunc = func(arguments ...interface{}) {
		var candle fyersTypes.FyersHistoricalCandle = arguments[0].(fyersTypes.FyersHistoricalCandle)
		callback(&candle)
	}

	listener := marketEventEmitter.AddListener(key, handlerFn)

	unsubscribeFn := func() {
		fmt.Println("stopping candle", key)
		marketEventEmitter.RemoveListener(key, listener)
	}

	return unsubscribeFn
}

func emitCandle(instrument string, timeFrame int, candle fyersTypes.FyersHistoricalCandle) {
	key := generateEventKey(marketConstants.MarketEventCandle, instrument, timeFrame)
	copyCandle := fyersTypes.FyersHistoricalCandle{}
	err := copier.Copy(&copyCandle, &candle)

	if err != nil {
		fmt.Println("Not able to copy candle for candle:", candle)
	}

	marketEventEmitter.EmitEvent(key, copyCandle)
}

func getPartialCandle(instrument string, timeFrame int) *fyersTypes.FyersHistoricalCandle {
	switch timeFrame {
	case marketConstants.TimeFrame15m:
		partialCandle15mMutex.RLock()
		candle, exists := partialCandle15m[instrument]
		if !exists {
			partialCandle15mMutex.RUnlock()
			return nil
		}
		var copyCandle *fyersTypes.FyersHistoricalCandle = &fyersTypes.FyersHistoricalCandle{}
		copier.Copy(copyCandle, candle)
		partialCandle15mMutex.RUnlock()
		return copyCandle
	}
	return nil
}

func setPartialCandle(instrument string, timeFrame int, candle *fyersTypes.FyersHistoricalCandle) {
	switch timeFrame {
	case marketConstants.TimeFrame15m:
		var copyCandle *fyersTypes.FyersHistoricalCandle = &fyersTypes.FyersHistoricalCandle{}
		copier.Copy(copyCandle, candle)
		partialCandle15mMutex.Lock()
		delete(partialCandle15m, instrument)
		partialCandle15m[instrument] = candle
		partialCandle15mMutex.Unlock()
	}
}

func updateCandle(instrument string, timeFrame int, tickData marketTypes.MarketTick) (*fyersTypes.FyersHistoricalCandle, *fyersTypes.FyersHistoricalCandle) {
	var lastCandle *fyersTypes.FyersHistoricalCandle = getPartialCandle(instrument, timeFrame)
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
			setPartialCandle(instrument, timeFrame, &newCandle)
		} else {
			// update same candle
			lastCandle.Close = ltp
			lastCandle.High = math.Max(lastCandle.High, ltp)
			lastCandle.Low = math.Min(lastCandle.Low, ltp)
			partialCandle = &fyersTypes.FyersHistoricalCandle{}
			copier.Copy(partialCandle, lastCandle)
			setPartialCandle(instrument, timeFrame, lastCandle)
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
		setPartialCandle(instrument, timeFrame, &firstCandle)
	}

	partialCandle.TS = tickData.TS
	partialCandle.DateString = utils.GetDateStringFromTimestamp(tickData.TS)
	partialCandle.Day = utils.GetWeekdayFromTimestamp(tickData.TS)

	return partialCandle, completeCandle
}

func FetchHistoricalData(instrument string, resolution string, fromDate time.Time, toDate time.Time, contFlag int) ([]fyersTypes.FyersHistoricalCandle, error) {
	var allCandles = make([]fyersTypes.FyersHistoricalCandle, 0)

	fmt.Println("fetching historical Data", instrument, resolution, fromDate.Local().Format(time.RFC3339), toDate.Local().Format(time.RFC3339), contFlag)

	currentDate := time.Time{}
	copier.Copy(&currentDate, &fromDate)

	for currentDate.Sub(toDate).Milliseconds() <= 0 {
		after3Months := currentDate.AddDate(0, 3, 0)
		if toDate.Sub(after3Months).Milliseconds() < 0 {
			after3Months = toDate
		}

		data, err := fyersSDK.FetchHistoricalData(instrument, resolution, currentDate.Unix(), after3Months.Unix(), contFlag)
		if err != nil {
			return make([]fyersTypes.FyersHistoricalCandle, 0), err
		}

		allCandles = append(allCandles, data...)

		currentDate = after3Months.AddDate(0, 0, 1)
	}

	sort.Slice(allCandles, func(i, j int) bool {
		return allCandles[i].TS < allCandles[j].TS
	})

	return allCandles, nil
}

func IsMarketWatchActive() bool {
	isMarketActiveMutex.RLock()
	active := isMarketActive
	isMarketActiveMutex.RUnlock()
	return active
}

func SetIsMarketWatchActive(active bool) {
	isMarketActiveMutex.Lock()
	isMarketActive = active
	isMarketActiveMutex.Unlock()
}

func IsWarmUpInProgress() bool {
	isWarmUpInProgressMutex.RLock()
	active := isWarmUpInProgress
	isWarmUpInProgressMutex.RUnlock()
	return active
}

func SetIsWarmUpInProgress(isWarmingUp bool) {
	isWarmUpInProgressMutex.Lock()
	isWarmUpInProgress = isWarmingUp
	isWarmUpInProgressMutex.Unlock()
}

func OnFyersWatchConnect() {
	// TODO: Pranav - if this is called while market is active then its an issue, handle it
	fmt.Println("connected to fyers socket")
	notifications.Broadcast(constants.AccessLevelAdmin, "Fyers Socket Connected")
}
func OnFyersWatchMessage(notification fyersWatchAPI.Notification) {
	fmt.Println("received message from fyers server", notification)
	fyersWatchNotificationChannel <- notification
}
func OnFyersWatchError(err error) {
	fmt.Println("error occured on fyers watch", err)
	notifications.Broadcast(constants.AccessLevelAdmin, "Error occured on fyers watch\n\n"+err.Error())
	Stop()
}
func OnFyersWatchDisconnect(err error) {
	fmt.Println("disconnected from fyers server", err, fyersWatchNotificationChannel)
	notifications.Broadcast(constants.AccessLevelAdmin, "Disconnected from fyers server\n\n"+err.Error())
	// will disconnect because of either error, or outside intervention
	// in both cases, market will be stopped by others
}

func GenerateFakeNotification(instrument string, ltp float64, timeStamp time.Time, totalBuyQty int64, totalSellQty int64) fyersWatchAPI.Notification {
	return fyersWatchAPI.Notification{
		SymbolData: fyersWatchAPI.SymbolDataNotification{
			Symbol:       instrument,
			Timestamp:    timeStamp,
			TotalBuyQty:  totalBuyQty,
			TotalSellQty: totalSellQty,
			Ltp:          float32(ltp),
		},
	}
}

func Start() (bool, error) {
	if IsWarmUpInProgress() {
		return false, fmt.Errorf("market is already warming up")
	}
	if IsMarketWatchActive() {
		return false, fmt.Errorf("market is already active")
	}

	SetIsWarmUpInProgress(true)
	defer SetIsWarmUpInProgress(false)

	partialCandle15mMutex.Lock()
	partialCandle15m = make(map[string]*fyersTypes.FyersHistoricalCandle)
	partialCandle15mMutex.Unlock()
	strategyIDToPositionalStrategyMap = make(map[primitive.ObjectID]*postionalStrategy.PositionalStrategy)
	strategyIDToTickSubscription = make(map[primitive.ObjectID]func())
	strategyIDToCandleSubscription = make(map[primitive.ObjectID]func())
	fyersWatchNotificationChannel = make(chan fyersWatchAPI.Notification)

	// Fetch all Instruments
	instrumentsCursor, err := instrumentModel.GetInstrumentCollection().Find(context.Background(), bson.M{})
	if err != nil {
		return false, err
	}
	instrumentSymbols := make([]string, 0)
	instruments := make([]instrumentModel.InstrumentModel, 0)
	for instrumentsCursor.Next(context.Background()) {
		instrument := instrumentModel.InstrumentModel{}
		instrumentsCursor.Decode(&instrument)
		instruments = append(instruments, instrument)
		instrumentSymbols = append(instrumentSymbols, instrument.Symbol)
	}
	instrumentValidityMap, err := fyersSDK.IsValidInstrumentMany(instrumentSymbols)
	if err != nil {
		return false, err
	}
	invalidInstrumentSymbols := make([]string, 0)
	for _, instrument := range instrumentSymbols {
		if valid := instrumentValidityMap[instrument]; !valid {
			invalidInstrumentSymbols = append(invalidInstrumentSymbols, instrument)
		}
	}
	if len(invalidInstrumentSymbols) > 0 {
		return false, fmt.Errorf("found invalid instruments, consider renaming them\n\n" + strings.Join(invalidInstrumentSymbols, "\n"))
	}
	if fyersSDK.GetFyersAccessToken() == "" {
		return false, fmt.Errorf("fyers access token does not exist, please login first")
	}

	_, err = fyersSDK.StartMarketWatch(instrumentSymbols, OnFyersWatchConnect, OnFyersWatchMessage, OnFyersWatchError, OnFyersWatchDisconnect)
	if err != nil {
		return false, err
	}
	instrumentIDToSymbolMap := make(map[primitive.ObjectID]string)
	for _, instrument := range instruments {
		instrumentIDToSymbolMap[instrument.ID] = instrument.Symbol
	}
	requestedInstruments := make(map[primitive.ObjectID]bool)

	// Positional Strategy
	positionalOpenTradesCursor, err := positionalTradeModel.GetPositionalTradeCollection().Find(context.Background(), bson.M{"status": marketConstants.TradeStatusOpen})
	if err != nil {
		OnFyersWatchError(err)
		return false, err
	}
	positionalOpenTrades := make([]positionalTradeModel.PositionalTrade, 0)
	positionalOpenTradesCursor.All(context.Background(), &positionalOpenTrades)
	positionalStrategyIDToOpenTradeMap := make(map[primitive.ObjectID]positionalTradeModel.PositionalTrade)
	for _, openTrade := range positionalOpenTrades {
		if _, exists := positionalStrategyIDToOpenTradeMap[openTrade.StrategyID]; exists {
			err := fmt.Errorf("should not have more than 1 open trade per strategy id, " + strategyConstants.StrategyPositional + " strategy id - " + openTrade.StrategyID.Hex() + ", trade id - " + openTrade.ID.Hex())
			OnFyersWatchError(err)
			return false, err
		}
		positionalStrategyIDToOpenTradeMap[openTrade.StrategyID] = openTrade
	}

	instrumentIDToPositionalStrategyListMap := make(map[primitive.ObjectID][]*postionalStrategy.PositionalStrategy)
	positionalStrategiesCursor, err := positionalStrategyModel.GetPositionalStrategyCollection().Find(context.Background(), bson.M{"isActive": true})
	if err != nil {
		OnFyersWatchError(err)
		return false, err
	}

	for positionalStrategiesCursor.Next(context.Background()) {
		strategy := positionalStrategyModel.PositionalStrategy{}
		positionalStrategiesCursor.Decode(&strategy)
		requestedInstruments[strategy.InstrumentID] = true

		pos := postionalStrategy.NewPositionalStrategy(instrumentIDToSymbolMap[strategy.InstrumentID])
		if openTrade, exists := positionalStrategyIDToOpenTradeMap[strategy.ID]; exists {
			pos.OpenTrade = &postionalStrategy.PositionalTrade{
				ID:            openTrade.ID,
				TradeType:     openTrade.TradeType,
				TradeTypeText: openTrade.TradeTypeText,
				Lots:          openTrade.Lots,
				Entry: &postionalStrategy.PositionalCandle{
					Candle: &fyersTypes.FyersHistoricalCandle{
						TS:         openTrade.Entry.Candle.TS,
						DateString: openTrade.Entry.Candle.DateString,
						Day:        openTrade.Entry.Candle.Day,
						Open:       openTrade.Entry.Candle.Open,
						High:       openTrade.Entry.Candle.High,
						Low:        openTrade.Entry.Candle.Low,
						Close:      openTrade.Entry.Candle.Close,
						Volume:     openTrade.Entry.Candle.Volume,
					},
					Indicators: &postionalStrategy.PositionalIndicators{
						SuperTrend: &indicatorTypes.SuperTrendData{
							Index:               openTrade.Entry.Indicators.SuperTrend.Index,
							ATR:                 openTrade.Entry.Indicators.SuperTrend.ATR,
							PastTRList:          openTrade.Entry.Indicators.SuperTrend.PastTRList,
							BasicUpperBound:     openTrade.Entry.Indicators.SuperTrend.BasicUpperBound,
							BasicLowerBound:     openTrade.Entry.Indicators.SuperTrend.BasicLowerBound,
							FinalUpperBound:     openTrade.Entry.Indicators.SuperTrend.FinalUpperBound,
							FinalLowerBound:     openTrade.Entry.Indicators.SuperTrend.FinalLowerBound,
							SuperTrend:          openTrade.Entry.Indicators.SuperTrend.SuperTrend,
							SuperTrendDirection: openTrade.Entry.Indicators.SuperTrend.SuperTrendDirection,
							IsUsable:            openTrade.Entry.Indicators.SuperTrend.IsUsable,
						},
					},
				},
				PL:             openTrade.PL,
				Exit:           nil,
				ExitReason:     marketConstants.TradeExitReasonNone,
				ExitReasonText: marketConstants.TradeExitReasonNoneText,
				Brokerage:      0.,
			}
		}
		pos.SetUserID(strategy.UserID)
		pos.SetID(strategy.ID)
		pos.SetOnCanEnter(func(ID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, tradeType int, candleClose float64) {
			if IsMarketWatchActive() && !IsWarmUpInProgress() {
				err := notifications.NotifyPositionalCanEnter(ID, timeFrame, instrument, userID, tradeType, candleClose)
				if err != nil {
					fmt.Println("not able to send notification onCanEnter", err)
				}
			}
		})
		pos.SetOnCanExit(func(ID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, forceExit int, candleClose float64, PL float64) {
			if IsMarketWatchActive() && !IsWarmUpInProgress() {
				notifications.NotifyPositionalCanExit(ID, timeFrame, instrument, userID, forceExit, candleClose, PL)
				if err != nil {
					fmt.Println("not able to send notification onCanExit", err)
				}
			}
		})
		unsubCandle := SubscribeCandle(instrumentIDToSymbolMap[strategy.InstrumentID], marketConstants.TextToTimeFrameMap[strategy.TimeFrame], pos.OnCandle)
		unsubTick := SubscribeTick(instrumentIDToSymbolMap[strategy.InstrumentID], marketConstants.TextToTimeFrameMap[strategy.TimeFrame], pos.OnTick)
		strategyIDToPositionalStrategyMap[strategy.ID] = pos

		if _, ok := instrumentIDToPositionalStrategyListMap[strategy.InstrumentID]; !ok {
			instrumentIDToPositionalStrategyListMap[strategy.InstrumentID] = make([]*postionalStrategy.PositionalStrategy, 0)
		}

		instrumentIDToPositionalStrategyListMap[strategy.InstrumentID] = append(instrumentIDToPositionalStrategyListMap[strategy.InstrumentID], pos)
		strategyIDToCandleSubscription[strategy.ID] = unsubCandle
		strategyIDToTickSubscription[strategy.ID] = unsubTick
	}

	fmt.Println(utils.BruteStringify(requestedInstruments))

	lastWarmUpTimestampOfInstrument := make(map[string]int64)
	lastWarmUpTimestampOfInstrumentMutex := &sync.Mutex{}

	toDate := time.Now()
	fromDate := toDate.Add(-marketConstants.WarmUpDuration)
	// For Testing:
	// toDate := time.Now().Add(-marketConstants.WarmUpDuration)
	// fromDate := toDate.Add(-marketConstants.WarmUpDuration)

	wg := &sync.WaitGroup{}
	var warmUpError error = nil
	for instrumentID := range requestedInstruments {
		wg.Add(1)
		fmt.Println(instrumentIDToSymbolMap[instrumentID], "symbol added")
		go func(instrumentID primitive.ObjectID, warmUpError *error) {
			candles, err := FetchHistoricalData(instrumentIDToSymbolMap[instrumentID], marketConstants.Resolution1m, fromDate, toDate, 0)
			if err != nil {
				if warmUpError == nil {
					OnFyersWatchError(err)
					warmUpError = &err
				}
				wg.Done()
				return
			}

			fmt.Println(len(candles), "candles")

			uniqueTimeFrames := make(map[int]bool)
			// And other strategies
			for _, pos := range instrumentIDToPositionalStrategyListMap[instrumentID] {
				uniqueTimeFrames[pos.TimeFrame] = true
			}

			instrument := instrumentIDToSymbolMap[instrumentID]
			for _, tick := range candles {
				lastWarmUpTimestampOfInstrumentMutex.Lock()
				lastWarmUpTimestampOfInstrument[instrument] = tick.TS
				lastWarmUpTimestampOfInstrumentMutex.Unlock()
				// fmt.Println("sending tick", tick)
				for timeFrame := range uniqueTimeFrames {
					partialCandle1, completeCandle1 := updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.Open, Volume: tick.Volume})
					// fmt.Println("1", completeCandle1)
					emitTick(instrument, timeFrame, *partialCandle1)
					if completeCandle1 != nil {
						emitCandle(instrument, timeFrame, *completeCandle1)
					}

					partialCandle2, completeCandle2 := updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.High, Volume: tick.Volume})
					// fmt.Println("2", completeCandle2)
					emitTick(instrument, timeFrame, *partialCandle2)
					if completeCandle2 != nil {
						emitCandle(instrument, timeFrame, *completeCandle2)
					}

					partialCandle3, completeCandle3 := updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.Low, Volume: tick.Volume})
					// fmt.Println("3", completeCandle3)
					emitTick(instrument, timeFrame, *partialCandle3)
					if completeCandle3 != nil {
						emitCandle(instrument, timeFrame, *completeCandle3)
					}

					partialCandle4, completeCandle4 := updateCandle(instrument, timeFrame, marketTypes.MarketTick{TS: tick.TS, LTP: tick.Close, Volume: tick.Volume})
					// fmt.Println("4", completeCandle4)
					emitTick(instrument, timeFrame, *partialCandle4)
					if completeCandle4 != nil {
						emitCandle(instrument, timeFrame, *completeCandle4)
					}
				}
			}
			wg.Done()
		}(instrumentID, &warmUpError)
	}
	wg.Wait()

	if warmUpError != nil {
		return false, warmUpError
	}

	// Call OnWarmUpComplete for all strategies
	for _, strategy := range strategyIDToPositionalStrategyMap {
		strategy.OnWarmUpComplete()
	}

	go func() {
		for notification := range fyersWatchNotificationChannel {
			fmt.Println("got notification", utils.BruteStringify(notification))
			symbolWithoutExchange := strings.Replace(notification.SymbolData.Symbol, fyersConstants.Exchange+":", "", 1)
			fmt.Println(symbolWithoutExchange)
			if lastWarmUpTimestampOfInstrument[symbolWithoutExchange] >= notification.SymbolData.Timestamp.Unix() {
				fmt.Println("skipping this tick because it came before lastWarmUpTimestampOfInstrument", symbolWithoutExchange, time.Unix(lastWarmUpTimestampOfInstrument[symbolWithoutExchange], 0), notification.SymbolData.Timestamp)
				continue
			}
			marketOpen := marketConstants.MarketOpenHours*60 + marketConstants.MarketOpenMinutes
			marketClose := marketConstants.MarketCloseHours*60 + marketConstants.MarketCloseMinutes
			tick := notification.SymbolData.Timestamp.Hour()*60 + notification.SymbolData.Timestamp.Minute()
			if tick < marketOpen || tick > marketClose {
				fmt.Println("skipping this tick because it is outside market hours", symbolWithoutExchange, time.Unix(lastWarmUpTimestampOfInstrument[symbolWithoutExchange], 0), notification.SymbolData.Timestamp)
				continue
			}
			for timeFrame := range marketConstants.TimeFrameToTextMap {
				if timeFrame == marketConstants.TimeFrameUnknown {
					continue
				}
				partialCandle, completeCandle := updateCandle(symbolWithoutExchange, timeFrame, marketTypes.MarketTick{TS: notification.SymbolData.Timestamp.Unix(), LTP: float64(notification.SymbolData.Ltp), Volume: float64(notification.SymbolData.TotalBuyQty + notification.SymbolData.TotalSellQty)})
				emitTick(symbolWithoutExchange, timeFrame, *partialCandle)
				if completeCandle != nil {
					emitCandle(symbolWithoutExchange, timeFrame, *completeCandle)
				}
			}
		}
	}()

	// For Testing:
	// go func() {
	// 	now := time.Now()
	// 	for instrumentID := range requestedInstruments {
	// 		candles, err := FetchHistoricalData(instrumentIDToSymbolMap[instrumentID], marketConstants.Resolution1m, toDate.Add(time.Hour*24), now, 0)
	// 		if err != nil {
	// 			OnFyersWatchError(err)
	// 			return
	// 		}

	// 		fmt.Println(len(candles), "candles")

	// 		instrument := instrumentIDToSymbolMap[instrumentID]
	// 		for _, tick := range candles {
	// 			if IsMarketWatchActive() {
	// 				OnFyersWatchMessage(GenerateFakeNotification(fyersConstants.Exchange+":"+instrument, tick.Open, time.Unix(tick.TS, 0), int64(tick.Volume), int64(tick.Volume)))
	// 				OnFyersWatchMessage(GenerateFakeNotification(fyersConstants.Exchange+":"+instrument, tick.Low, time.Unix(tick.TS, 0), int64(tick.Volume), int64(tick.Volume)))
	// 				OnFyersWatchMessage(GenerateFakeNotification(fyersConstants.Exchange+":"+instrument, tick.High, time.Unix(tick.TS, 0), int64(tick.Volume), int64(tick.Volume)))
	// 				OnFyersWatchMessage(GenerateFakeNotification(fyersConstants.Exchange+":"+instrument, tick.Close, time.Unix(tick.TS, 0), int64(tick.Volume), int64(tick.Volume)))
	// 				time.Sleep(time.Second)
	// 			}
	// 		}
	// 	}
	// }()

	SetIsMarketWatchActive(true)
	return true, nil
}

func Stop() (bool, error) {
	if IsWarmUpInProgress() {
		return false, fmt.Errorf("market is currently warming up, please wait for it to finish")
	}
	if !IsMarketWatchActive() {
		return false, fmt.Errorf("market is already inactive")
	}

	_, err := fyersSDK.StopMarketWatch()
	if err != nil {
		return false, err
	}

	marketEventEmitter.RemoveAllListeners()
	close(fyersWatchNotificationChannel)

	SetIsMarketWatchActive(false)

	return true, nil
}

func UnsubscribeStrategy(strategyID primitive.ObjectID) {
	unsubCandle := strategyIDToCandleSubscription[strategyID]
	if unsubCandle != nil {
		unsubCandle()
		delete(strategyIDToCandleSubscription, strategyID)
	}
	unsubTick := strategyIDToTickSubscription[strategyID]
	if unsubTick != nil {
		unsubTick()
		delete(strategyIDToTickSubscription, strategyID)
	}
}

func EnterPositionalTrade(strategyID primitive.ObjectID, price float64, lots int, forceTradeType int, entryAt time.Time) (bool, error) {
	pos, exists := strategyIDToPositionalStrategyMap[strategyID]
	if !exists {
		return false, fmt.Errorf("invalid strategy id")
	}

	openTrade, err := pos.Enter(price, lots, entryAt, forceTradeType)
	if err != nil {
		return false, err
	}

	openTradeObj := positionalTradeModel.PositionalTrade{
		ID:            openTrade.ID,
		StrategyID:    strategyID,
		Status:        marketConstants.TradeStatusOpen,
		StatusText:    marketConstants.TradeStatusOpenText,
		TradeType:     forceTradeType,
		TradeTypeText: marketConstants.TradeTypeToTextMap[forceTradeType],
		Lots:          lots,
		Entry: positionalTradeModel.PositionalCombinedCandle{
			Candle: positionalTradeModel.PositionalCandle{
				TS:         openTrade.Entry.Candle.TS,
				DateString: openTrade.Entry.Candle.DateString,
				Day:        openTrade.Entry.Candle.Day,
				Open:       openTrade.Entry.Candle.Open,
				High:       openTrade.Entry.Candle.High,
				Low:        openTrade.Entry.Candle.Low,
				Close:      openTrade.Entry.Candle.Close,
				Volume:     openTrade.Entry.Candle.Volume,
			},
			Indicators: positionalTradeModel.PositionalIndicators{
				SuperTrend: positionalTradeModel.PositionalSuperTrendData{
					Index:               openTrade.Entry.Indicators.SuperTrend.Index,
					ATR:                 openTrade.Entry.Indicators.SuperTrend.ATR,
					PastTRList:          openTrade.Entry.Indicators.SuperTrend.PastTRList,
					BasicUpperBound:     openTrade.Entry.Indicators.SuperTrend.BasicUpperBound,
					BasicLowerBound:     openTrade.Entry.Indicators.SuperTrend.BasicLowerBound,
					FinalUpperBound:     openTrade.Entry.Indicators.SuperTrend.FinalUpperBound,
					FinalLowerBound:     openTrade.Entry.Indicators.SuperTrend.FinalLowerBound,
					SuperTrend:          openTrade.Entry.Indicators.SuperTrend.SuperTrend,
					SuperTrendDirection: openTrade.Entry.Indicators.SuperTrend.SuperTrendDirection,
					IsUsable:            openTrade.Entry.Indicators.SuperTrend.IsUsable,
				},
			},
		},
		PL:             openTrade.PL,
		Brokerage:      openTrade.Brokerage,
		ExitReason:     openTrade.ExitReason,
		ExitReasonText: openTrade.ExitReasonText,
		UpdatedAt:      time.Now().Unix(),
	}

	delResult := positionalTradeModel.GetPositionalTradeCollection().FindOneAndDelete(context.Background(), bson.M{
		"strategyID": strategyID,
		"status":     marketConstants.TradeStatusOpen,
	})

	err = delResult.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		// what else could it be?
		return false, err
	}

	_, err = positionalTradeModel.GetPositionalTradeCollection().InsertOne(context.Background(), openTradeObj)
	if err != nil {
		return false, err
	}

	return true, nil
}

func ExitPositionalTrade(strategyID primitive.ObjectID, price float64, forceExit bool, exitAt time.Time) (bool, error) {
	pos, exists := strategyIDToPositionalStrategyMap[strategyID]
	if !exists {
		return false, fmt.Errorf("invalid strategy id")
	}

	closedTrade, err := pos.Exit(price, forceExit, exitAt)
	if err != nil {
		return false, err
	}

	getResult := positionalTradeModel.
		GetPositionalTradeCollection().
		FindOne(context.Background(), bson.M{
			"strategyID": strategyID,
			"status":     marketConstants.TradeStatusOpen,
		})

	err = getResult.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, fmt.Errorf("open trade does not exist in the db while finding")
		}
		return false, err
	}

	openTrade := positionalTradeModel.PositionalTrade{}
	getResult.Decode(&openTrade)

	_, err = positionalTradeModel.
		GetPositionalTradeCollection().
		DeleteOne(context.Background(), bson.M{"_id": openTrade.ID})

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, fmt.Errorf("open trade does not exist in the db while deleting")
		}
		return false, err
	}

	openTrade.Exit = positionalTradeModel.PositionalCombinedCandle{
		Candle: positionalTradeModel.PositionalCandle{
			TS:         closedTrade.Exit.Candle.TS,
			DateString: closedTrade.Exit.Candle.DateString,
			Day:        closedTrade.Exit.Candle.Day,
			Open:       closedTrade.Exit.Candle.Open,
			High:       closedTrade.Exit.Candle.High,
			Low:        closedTrade.Exit.Candle.Low,
			Close:      closedTrade.Exit.Candle.Close,
			Volume:     closedTrade.Exit.Candle.Volume,
		},
		Indicators: positionalTradeModel.PositionalIndicators{
			SuperTrend: positionalTradeModel.PositionalSuperTrendData{
				Index:               closedTrade.Exit.Indicators.SuperTrend.Index,
				ATR:                 closedTrade.Exit.Indicators.SuperTrend.ATR,
				PastTRList:          closedTrade.Exit.Indicators.SuperTrend.PastTRList,
				BasicUpperBound:     closedTrade.Exit.Indicators.SuperTrend.BasicUpperBound,
				BasicLowerBound:     closedTrade.Exit.Indicators.SuperTrend.BasicLowerBound,
				FinalUpperBound:     closedTrade.Exit.Indicators.SuperTrend.FinalUpperBound,
				FinalLowerBound:     closedTrade.Exit.Indicators.SuperTrend.FinalLowerBound,
				SuperTrend:          closedTrade.Exit.Indicators.SuperTrend.SuperTrend,
				SuperTrendDirection: closedTrade.Exit.Indicators.SuperTrend.SuperTrendDirection,
				IsUsable:            closedTrade.Exit.Indicators.SuperTrend.IsUsable,
			},
		},
	}

	openTrade.Status = marketConstants.TradeStatusClosed
	openTrade.StatusText = marketConstants.TradeStatusClosedText
	openTrade.PL = closedTrade.PL
	openTrade.Brokerage = closedTrade.Brokerage
	openTrade.ExitReason = closedTrade.ExitReason
	openTrade.ExitReasonText = closedTrade.ExitReasonText

	_, err = positionalTradeModel.
		GetPositionalTradeCollection().
		InsertOne(context.Background(), openTrade)

	if err != nil {
		return false, err
	}

	return true, nil
}

func GetPositionalOpenTrades(userID primitive.ObjectID) ([]*postionalStrategy.PositionalTrade, []*postionalStrategy.PositionalStrategy, error) {
	openTradeList := make([]*postionalStrategy.PositionalTrade, 0)
	strategyList := make([]*postionalStrategy.PositionalStrategy, 0)
	posStrategyCursor, err := positionalStrategyModel.GetPositionalStrategyCollection().
		Find(context.Background(), bson.M{
			"userID":   userID,
			"isActive": true,
		})
	if err != nil {
		return make([]*postionalStrategy.PositionalTrade, 0), make([]*postionalStrategy.PositionalStrategy, 0), err
	}
	for posStrategyCursor.Next(context.Background()) {
		posStrategy := positionalStrategyModel.PositionalStrategy{}
		posStrategyCursor.Decode(&posStrategy)
		fmt.Println(utils.BruteStringify(posStrategy))
		strategyID := posStrategy.ID
		strategy, exists := strategyIDToPositionalStrategyMap[strategyID]
		if !exists {
			return make([]*postionalStrategy.PositionalTrade, 0), make([]*postionalStrategy.PositionalStrategy, 0), fmt.Errorf("system does not have knowledge of this " + strategyConstants.StrategyPositional + " strategy - " + strategyID.Hex())
		}

		openTrade := strategy.GetOpenTrade()
		if openTrade != nil {
			openTradeList = append(openTradeList, openTrade)
			strategyList = append(strategyList, strategy)
		}
	}

	return openTradeList, strategyList, nil
}

func GetAllPositionalOpenTrades() ([]*postionalStrategy.PositionalTrade, []*postionalStrategy.PositionalStrategy, error) {
	openTradeList := make([]*postionalStrategy.PositionalTrade, 0)
	strategyList := make([]*postionalStrategy.PositionalStrategy, 0)
	posStrategyCursor, err := positionalStrategyModel.GetPositionalStrategyCollection().
		Find(context.Background(), bson.M{
			"isActive": true,
		})
	if err != nil {
		return make([]*postionalStrategy.PositionalTrade, 0), make([]*postionalStrategy.PositionalStrategy, 0), err
	}
	for posStrategyCursor.Next(context.Background()) {
		posStrategy := positionalStrategyModel.PositionalStrategy{}
		posStrategyCursor.Decode(&posStrategy)
		fmt.Println(utils.BruteStringify(posStrategy))
		strategyID := posStrategy.ID
		strategy, exists := strategyIDToPositionalStrategyMap[strategyID]
		if !exists {
			return make([]*postionalStrategy.PositionalTrade, 0), make([]*postionalStrategy.PositionalStrategy, 0), fmt.Errorf("system does not have knowledge of this " + strategyConstants.StrategyPositional + " strategy - " + strategyID.Hex())
		}

		openTrade := strategy.GetOpenTrade()
		if openTrade != nil {
			openTradeList = append(openTradeList, openTrade)
			strategyList = append(strategyList, strategy)
		}
	}

	return openTradeList, strategyList, nil
}

func RemovePositionalStrategy(strategyID primitive.ObjectID) {
	UnsubscribeStrategy(strategyID)
	delete(strategyIDToPositionalStrategyMap, strategyID)
}
