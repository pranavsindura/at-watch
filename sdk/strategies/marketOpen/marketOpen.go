package marketOpenStrategy

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	marketOpenStrategyConstants "github.com/pranavsindura/at-watch/constants/strategies/marketOpen"
	backtestSDK "github.com/pranavsindura/at-watch/sdk/backtest"
	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	"github.com/pranavsindura/at-watch/utils"
	marketUtils "github.com/pranavsindura/at-watch/utils/market"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MarketOpenConfig struct {
}

type MarketOpenCandle struct {
	Index  int                               `json:"index"`
	Candle *fyersTypes.FyersHistoricalCandle `json:"candle"`
}

type MarketOpenTrade struct {
	ID             primitive.ObjectID `json:"id"`
	TradeType      int                `json:"tradeType"`
	TradeTypeText  string             `json:"tradeTypeText"`
	Lots           int                `json:"lots"`
	Target         float64            `json:"target"`
	StopLoss       float64            `json:"stopLoss"`
	Entry          *MarketOpenCandle  `json:"entry"`
	PL             float64            `json:"PL"`
	Exit           *MarketOpenCandle  `json:"exit"`
	ExitReason     int                `json:"exitReason"`
	ExitReasonText string             `json:"exitReasonText"`
	Brokerage      float64            `json:"brokerage"`
	UpdatedAtTS    int64              `json:"updatedAtTS"`
	UpdatedAtLTP   float64            `json:"updatedAtLTP"`
}

type MarketOpenStrategy struct {
	UserID                   primitive.ObjectID `json:"userID"`
	ID                       primitive.ObjectID `json:"id"`
	Instrument               string             `json:"instrument"`
	TimeFrame                int                `json:"timeFrame"`
	OpenTrade                *MarketOpenTrade   `json:"openTrade"`
	ClosedTrades             []*MarketOpenTrade `json:"closedTrades"`
	WaitingToOpenTrade       bool               `json:"waitingToOpenTrade"`
	WaitingToCloseTrade      bool               `json:"waitingToCloseTrade"`
	LastCandle               *MarketOpenCandle  `json:"lastCandle"`
	FirstCandleOfTheDay      *MarketOpenCandle  `json:"firstCandleOfTheDay"`
	Config                   MarketOpenConfig   `json:"config"`
	openTradeMutex           *sync.RWMutex
	closedTradesMutex        *sync.RWMutex
	waitingToOpenTradeMutex  *sync.RWMutex
	waitingToCloseTradeMutex *sync.RWMutex
	lastCandleMutex          *sync.RWMutex
	firstCandleOfTheDayMutex *sync.RWMutex
	IsBacktestModeEnabled    bool
	onCanEnter               func(ID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, tradeType int, candleClose float64)
	onCanExit                func(ID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, exitReason int, candleClose float64, PL float64)
}

func NewMarketOpenStrategy(instrument string) *MarketOpenStrategy {
	strategy := &MarketOpenStrategy{
		Instrument:               instrument,
		TimeFrame:                marketOpenStrategyConstants.TimeFrame,
		OpenTrade:                nil,
		ClosedTrades:             make([]*MarketOpenTrade, 0),
		WaitingToOpenTrade:       false,
		WaitingToCloseTrade:      false,
		LastCandle:               nil,
		FirstCandleOfTheDay:      nil,
		Config:                   MarketOpenConfig{},
		openTradeMutex:           &sync.RWMutex{},
		closedTradesMutex:        &sync.RWMutex{},
		waitingToOpenTradeMutex:  &sync.RWMutex{},
		waitingToCloseTradeMutex: &sync.RWMutex{},
		lastCandleMutex:          &sync.RWMutex{},
		firstCandleOfTheDayMutex: &sync.RWMutex{},
		IsBacktestModeEnabled:    false,
		onCanEnter:               nil,
		onCanExit:                nil,
	}
	return strategy
}

func (pos *MarketOpenStrategy) generateTradeId() primitive.ObjectID {
	return primitive.NewObjectID()
}

func (pos *MarketOpenStrategy) SetIsBacktestModeEnabled(enabled bool) {
	pos.IsBacktestModeEnabled = enabled
}

func (pos *MarketOpenStrategy) SetUserID(userID primitive.ObjectID) {
	pos.UserID = userID
}

func (pos *MarketOpenStrategy) SetID(ID primitive.ObjectID) {
	pos.ID = ID
}

func (pos *MarketOpenStrategy) SetOnCanEnter(onCanEnter func(ID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, tradeType int, candleClose float64)) {
	pos.onCanEnter = onCanEnter
}

func (pos *MarketOpenStrategy) SetOnCanExit(onCanExit func(ID primitive.ObjectID, timeFrame int, instrument string, userID primitive.ObjectID, exitReason int, candleClose float64, PL float64)) {
	pos.onCanExit = onCanExit
}

func (pos *MarketOpenStrategy) getTradeType(candle *MarketOpenCandle, firstCandleOfTheDay *MarketOpenCandle) int {
	// candle will always be non nil but
	// firstCandleOfTheDay can be nil if candle is the first candle
	if firstCandleOfTheDay == nil {
		return marketConstants.TradeTypeNone
	}

	diff := time.Unix(candle.Candle.TS, 0).Sub(time.Unix(firstCandleOfTheDay.Candle.TS, 0))
	// check only marketOpenStrategyConstants.CheckCandleCount after the first candle
	if diff > marketOpenStrategyConstants.TimeFrameDuration*time.Duration(marketOpenStrategyConstants.CheckCandleCount+1) {
		return marketConstants.TradeTypeNone
	}

	if candle.Candle.Close > firstCandleOfTheDay.Candle.High {
		return marketConstants.TradeTypeBuy
	}
	if candle.Candle.Close < firstCandleOfTheDay.Candle.Low {
		return marketConstants.TradeTypeSell
	}

	return marketConstants.TradeTypeNone
}

func (pos *MarketOpenStrategy) canEnter(candle *MarketOpenCandle, firstCandleOfTheDay *MarketOpenCandle) bool {
	return pos.getTradeType(candle, firstCandleOfTheDay) != marketConstants.TradeTypeNone
}

func (pos *MarketOpenStrategy) enter(candle *MarketOpenCandle, firstCandleOfTheDay *MarketOpenCandle, lots int, forceTradeType int) *MarketOpenTrade {
	copyCandle := &MarketOpenCandle{}
	copier.CopyWithOption(copyCandle, candle, copier.Option{DeepCopy: true})
	var trade *MarketOpenTrade = nil
	var tradeType int
	if forceTradeType != marketConstants.TradeTypeNone {
		tradeType = forceTradeType
	} else {
		tradeType = pos.getTradeType(copyCandle, firstCandleOfTheDay)
	}

	switch tradeType {
	case marketConstants.TradeTypeBuy:
		trade = &MarketOpenTrade{
			ID:             pos.generateTradeId(),
			TradeType:      tradeType,
			TradeTypeText:  marketConstants.TradeTypeBuyText,
			Lots:           lots,
			Entry:          copyCandle,
			PL:             0.,
			Exit:           nil,
			ExitReason:     marketConstants.TradeExitReasonNone,
			ExitReasonText: marketConstants.TradeExitReasonNoneText,
			Brokerage:      0.,
			UpdatedAtTS:    candle.Candle.TS,
			UpdatedAtLTP:   candle.Candle.Close,
			Target:         candle.Candle.Close + marketOpenStrategyConstants.TargetBuffer,
			StopLoss:       candle.Candle.Close - marketOpenStrategyConstants.StopLossBuffer,
		}
	case marketConstants.TradeTypeSell:
		trade = &MarketOpenTrade{
			ID:             pos.generateTradeId(),
			TradeType:      tradeType,
			TradeTypeText:  marketConstants.TradeTypeSellText,
			Lots:           lots,
			Entry:          copyCandle,
			PL:             0.,
			Exit:           nil,
			ExitReason:     marketConstants.TradeExitReasonNone,
			ExitReasonText: marketConstants.TradeExitReasonNoneText,
			Brokerage:      0.,
			UpdatedAtTS:    candle.Candle.TS,
			UpdatedAtLTP:   candle.Candle.Close,
			Target:         candle.Candle.Close - marketOpenStrategyConstants.TargetBuffer,
			StopLoss:       candle.Candle.Close + marketOpenStrategyConstants.StopLossBuffer,
		}
	}

	return trade
}

func (pos *MarketOpenStrategy) getExitReason(candle *MarketOpenCandle, trade *MarketOpenTrade) int {
	switch trade.TradeType {
	case marketConstants.TradeTypeBuy:
		if candle.Candle.Close >= trade.Target {
			return marketConstants.TradeExitReasonTargetReached
		}
		if candle.Candle.Close <= trade.StopLoss {
			return marketConstants.TradeExitReasonStopLossHit
		}
		cutOffMinutes := marketOpenStrategyConstants.CutOffHour*60 + marketOpenStrategyConstants.CutOffMinute
		currentMinutes := time.Unix(candle.Candle.TS, 0).Hour()*60 + time.Unix(candle.Candle.TS, 0).Minute()
		if currentMinutes >= cutOffMinutes {
			return marketConstants.TradeExitReasonSystemForceExit
		}
		return marketConstants.TradeExitReasonNone
	case marketConstants.TradeTypeSell:
		if candle.Candle.Close <= trade.Target {
			return marketConstants.TradeExitReasonTargetReached
		}
		if candle.Candle.Close >= trade.StopLoss {
			return marketConstants.TradeExitReasonStopLossHit
		}
		cutOffMinutes := marketOpenStrategyConstants.CutOffHour*60 + marketOpenStrategyConstants.CutOffMinute
		currentMinutes := time.Unix(candle.Candle.TS, 0).Hour()*60 + time.Unix(candle.Candle.TS, 0).Minute()
		if currentMinutes >= cutOffMinutes {
			return marketConstants.TradeExitReasonSystemForceExit
		}
		return marketConstants.TradeExitReasonNone
	default:
		return marketConstants.TradeExitReasonNone
	}
}

func (pos *MarketOpenStrategy) canExit(candle *MarketOpenCandle, trade *MarketOpenTrade) bool {
	return pos.getExitReason(candle, trade) != marketConstants.TradeExitReasonNone
}

func (pos *MarketOpenStrategy) exit(candle *MarketOpenCandle, trade *MarketOpenTrade, forceExit bool) *MarketOpenTrade {
	copyCandle := &MarketOpenCandle{}
	copier.CopyWithOption(copyCandle, candle, copier.Option{DeepCopy: true})

	var exitReason int
	if forceExit {
		exitReason = marketConstants.TradeExitReasonUserForceExit
	} else {
		exitReason = pos.getExitReason(candle, trade)
	}

	var completedTrade *MarketOpenTrade = nil

	switch trade.TradeType {
	case marketConstants.TradeTypeBuy:
		PL := (candle.Candle.Close - trade.Entry.Candle.Close) * float64(trade.Lots)

		completedTrade = &MarketOpenTrade{}
		copier.CopyWithOption(completedTrade, trade, copier.Option{DeepCopy: true})
		completedTrade.PL = PL
		completedTrade.Exit = copyCandle
		completedTrade.ExitReason = exitReason
		completedTrade.ExitReasonText = marketConstants.TradeExitReasonToTextMap[exitReason]
		completedTrade.Brokerage = marketUtils.CalculateZerodhaNSEOptionsCharges(trade.Entry.Candle.Close, candle.Candle.Close, trade.Lots)
	case marketConstants.TradeTypeSell:
		PL := -(candle.Candle.Close - trade.Entry.Candle.Close) * float64(trade.Lots)

		completedTrade = &MarketOpenTrade{}
		copier.CopyWithOption(completedTrade, trade, copier.Option{DeepCopy: true})
		completedTrade.PL = PL
		completedTrade.Exit = copyCandle
		completedTrade.ExitReason = exitReason
		completedTrade.ExitReasonText = marketConstants.TradeExitReasonToTextMap[exitReason]
		completedTrade.Brokerage = marketUtils.CalculateZerodhaNSEOptionsCharges(trade.Entry.Candle.Close, candle.Candle.Close, trade.Lots)
	}

	return completedTrade
}

func (pos *MarketOpenStrategy) update(candle *MarketOpenCandle, trade *MarketOpenTrade) *MarketOpenTrade {
	var updatedTrade *MarketOpenTrade = &MarketOpenTrade{}
	copier.CopyWithOption(updatedTrade, trade, copier.Option{DeepCopy: true})

	switch trade.TradeType {
	case marketConstants.TradeTypeBuy:
		{
			PL := (candle.Candle.Close - trade.Entry.Candle.Close) * float64(trade.Lots)
			updatedTrade.PL = PL
		}
	case marketConstants.TradeTypeSell:
		{
			PL := -(candle.Candle.Close - trade.Entry.Candle.Close) * float64(trade.Lots)
			updatedTrade.PL = PL
		}
	}

	return updatedTrade
}

func (pos *MarketOpenStrategy) onCandle(candleData *MarketOpenCandle) {
	// pos.waitingToCloseTradeMutex.Lock()
	// pos.WaitingToCloseTrade = false
	// pos.waitingToCloseTradeMutex.Unlock()

	// pos.openTradeMutex.RLock()
	// if pos.OpenTrade != nil && pos.canExit(candleData, pos.OpenTrade) {
	// 	pos.waitingToCloseTradeMutex.Lock()
	// 	pos.WaitingToCloseTrade = true
	// 	pos.waitingToCloseTradeMutex.Unlock()
	// 	if pos.onCanExit != nil {
	// 		pos.onCanExit(pos.ID, pos.TimeFrame, pos.Instrument, pos.UserID, pos.getExitReason(candleData, pos.OpenTrade), candleData.Candle.Close, pos.OpenTrade.PL)
	// 	}
	// }
	// pos.openTradeMutex.RUnlock()
}

func (pos *MarketOpenStrategy) onTick(candleData *MarketOpenCandle) {
	pos.waitingToOpenTradeMutex.Lock()
	pos.WaitingToOpenTrade = false
	pos.waitingToOpenTradeMutex.Unlock()

	pos.waitingToCloseTradeMutex.Lock()
	pos.WaitingToCloseTrade = false
	pos.waitingToCloseTradeMutex.Unlock()

	pos.openTradeMutex.Lock()
	if pos.OpenTrade != nil {
		pos.OpenTrade = pos.update(candleData, pos.OpenTrade)
	}
	pos.openTradeMutex.Unlock()

	pos.openTradeMutex.RLock()
	if pos.OpenTrade != nil && pos.canExit(candleData, pos.OpenTrade) {
		pos.waitingToCloseTradeMutex.Lock()
		pos.WaitingToCloseTrade = true
		pos.waitingToCloseTradeMutex.Unlock()
		if pos.onCanExit != nil {
			pos.onCanExit(pos.ID, pos.TimeFrame, pos.Instrument, pos.UserID, pos.getExitReason(candleData, pos.OpenTrade), candleData.Candle.Close, pos.OpenTrade.PL)
		}
	}

	pos.firstCandleOfTheDayMutex.RLock()
	if pos.canEnter(candleData, pos.FirstCandleOfTheDay) && (pos.OpenTrade == nil || pos.OpenTrade.TradeType != pos.getTradeType(candleData, pos.FirstCandleOfTheDay)) {
		// fmt.Println("waiting to enter trade, but must exit existing trade if any", marketConstants.TradeTypeToTextMap[pos.getTradeType(candleData)])
		pos.waitingToOpenTradeMutex.Lock()
		pos.WaitingToOpenTrade = true
		pos.waitingToOpenTradeMutex.Unlock()
		if pos.onCanEnter != nil {
			pos.onCanEnter(pos.ID, pos.TimeFrame, pos.Instrument, pos.UserID, pos.getTradeType(candleData, pos.FirstCandleOfTheDay), candleData.Candle.Close)
		}
	}
	pos.firstCandleOfTheDayMutex.RUnlock()
	pos.openTradeMutex.RUnlock()
}

// This function creates a MarketOpenCandle with updated indicators using its last known candle and a FyersHistoricalCandle
func (pos *MarketOpenStrategy) createCandle(candle *fyersTypes.FyersHistoricalCandle) *MarketOpenCandle {
	lastCandleIndex := 0
	pos.lastCandleMutex.RLock()
	if pos.LastCandle != nil {
		lastCandleIndex = pos.LastCandle.Index
	}
	pos.lastCandleMutex.RUnlock()
	copyCandle := &fyersTypes.FyersHistoricalCandle{}
	copier.CopyWithOption(copyCandle, candle, copier.Option{DeepCopy: true})

	if pos.LastCandle != nil && time.Unix(candle.TS, 0).Sub(time.Unix(pos.LastCandle.Candle.TS, 0)) > marketOpenStrategyConstants.TimeFrameDuration {
		lastCandleIndex++
	}

	posCandle := &MarketOpenCandle{
		Candle: copyCandle,
		Index:  lastCandleIndex + 1,
	}

	return posCandle
}

func (pos *MarketOpenStrategy) OnCandle(candle *fyersTypes.FyersHistoricalCandle) {
	posCandle := pos.createCandle(candle)
	// pos.lastCandleMutex.Lock()
	// pos.LastCandle = &MarketOpenCandle{}
	// copier.CopyWithOption(pos.LastCandle, posCandle, copier.Option{DeepCopy: true})
	// pos.lastCandleMutex.Unlock()

	pos.firstCandleOfTheDayMutex.Lock()
	if pos.FirstCandleOfTheDay == nil || // if FirstCandleOfTheDay does not exist
		time.Unix(candle.TS, 0).Sub(time.Unix(pos.FirstCandleOfTheDay.Candle.TS, 0)) > time.Hour*12 { // if the times are more than 12 hours apart then its a different day
		pos.FirstCandleOfTheDay = &MarketOpenCandle{}
		copier.CopyWithOption(pos.FirstCandleOfTheDay, posCandle, copier.Option{DeepCopy: true})
	}
	pos.firstCandleOfTheDayMutex.Unlock()

	pos.onCandle(posCandle)
}

func (pos *MarketOpenStrategy) OnTick(candle *fyersTypes.FyersHistoricalCandle) {
	pos.firstCandleOfTheDayMutex.RLock()
	if pos.FirstCandleOfTheDay == nil || // if FirstCandleOfTheDay does not exist
		time.Unix(candle.TS, 0).Sub(time.Unix(pos.FirstCandleOfTheDay.Candle.TS, 0)) > time.Hour*12 { // if the times are more than 12 hours apart then its a different day
		// then we wait for FirstCandleOfTheDay to be set by OnCandle
		pos.firstCandleOfTheDayMutex.RUnlock()
		return
	}
	pos.firstCandleOfTheDayMutex.RUnlock()

	// we will wait for firstCandleOfTheDay to be valid before doing anything with the ticks
	posCandle := pos.createCandle(candle)
	pos.lastCandleMutex.Lock()
	pos.LastCandle = &MarketOpenCandle{}
	copier.CopyWithOption(pos.LastCandle, posCandle, copier.Option{DeepCopy: true})
	pos.lastCandleMutex.Unlock()

	pos.onTick(posCandle)
}

func (pos *MarketOpenStrategy) Enter(price float64, lots int, entryAt time.Time, forceTradeType int) (*MarketOpenTrade, error) {
	pos.waitingToOpenTradeMutex.RLock()
	pos.waitingToCloseTradeMutex.RLock()
	pos.openTradeMutex.RLock()
	if !pos.WaitingToOpenTrade || pos.WaitingToCloseTrade || pos.OpenTrade != nil {
		pos.waitingToOpenTradeMutex.RUnlock()
		pos.waitingToCloseTradeMutex.RUnlock()
		pos.openTradeMutex.RUnlock()
		return nil, fmt.Errorf("strategy is not waiting to open trade, or there is a trade waiting to close, or there is already an open trade in the system")
	}
	pos.waitingToOpenTradeMutex.RUnlock()
	pos.waitingToCloseTradeMutex.RUnlock()
	pos.openTradeMutex.RUnlock()
	/*
		If i am waiting to open trade, last candle must be present
		as, it is only set true by OnCandle, which will set the last candle
		If we persist states, we must pass the system thru a backtest phase
		so that last candle is set
	*/
	pos.lastCandleMutex.RLock()
	candle := &fyersTypes.FyersHistoricalCandle{
		TS:         entryAt.Unix(),
		Open:       pos.LastCandle.Candle.Open,
		Close:      price,
		High:       pos.LastCandle.Candle.High,
		Low:        pos.LastCandle.Candle.Low,
		Volume:     pos.LastCandle.Candle.Volume,
		DateString: utils.GetDateStringFromTimestamp(entryAt.Unix()),
		Day:        utils.GetWeekdayFromTimestamp(entryAt.Unix()),
	}
	pos.lastCandleMutex.RUnlock()
	posCandle := pos.createCandle(candle)

	pos.firstCandleOfTheDayMutex.RLock()
	trade := pos.enter(posCandle, pos.FirstCandleOfTheDay, lots, forceTradeType)
	if trade != nil {
		pos.openTradeMutex.Lock()
		pos.OpenTrade = &MarketOpenTrade{}
		copier.CopyWithOption(pos.OpenTrade, trade, copier.Option{DeepCopy: true})
		pos.openTradeMutex.Unlock()

		pos.waitingToOpenTradeMutex.Lock()
		pos.WaitingToOpenTrade = false
		pos.waitingToOpenTradeMutex.Unlock()
	}
	pos.firstCandleOfTheDayMutex.RUnlock()

	if trade == nil {
		return nil, fmt.Errorf("something went wrong")
	}

	return trade, nil
}

func (pos *MarketOpenStrategy) Exit(price float64, forceExit bool, exitAt time.Time) (*MarketOpenTrade, error) {
	pos.openTradeMutex.RLock()
	pos.waitingToCloseTradeMutex.RLock()
	if !((pos.WaitingToCloseTrade || forceExit) && pos.OpenTrade != nil) {
		pos.openTradeMutex.RUnlock()
		pos.waitingToCloseTradeMutex.RUnlock()
		return nil, fmt.Errorf("strategy is not waiting to close a trade or there is no open trade in the system")
	}
	pos.openTradeMutex.RUnlock()
	pos.waitingToCloseTradeMutex.RUnlock()
	/*
		If i am waiting to close trade, last candle must be present
		as, it is only set true by OnCandle, which will set the last candle
		If we persist states, we must pass the system thru a backtest phase
		so that last candle is set
	*/
	pos.lastCandleMutex.RLock()
	candle := &fyersTypes.FyersHistoricalCandle{
		TS:         exitAt.Unix(),
		Open:       pos.LastCandle.Candle.Open,
		Close:      price,
		High:       pos.LastCandle.Candle.High,
		Low:        pos.LastCandle.Candle.Low,
		Volume:     pos.LastCandle.Candle.Volume,
		DateString: utils.GetDateStringFromTimestamp(exitAt.Unix()),
		Day:        utils.GetWeekdayFromTimestamp(exitAt.Unix()),
	}
	pos.lastCandleMutex.RUnlock()
	posCandle := pos.createCandle(candle)
	pos.openTradeMutex.RLock()
	trade := pos.exit(posCandle, pos.OpenTrade, forceExit)
	pos.openTradeMutex.RUnlock()
	if trade != nil {
		if pos.IsBacktestModeEnabled {
			pos.ClosedTrades = append(pos.ClosedTrades, trade)
		}
		pos.openTradeMutex.Lock()
		pos.OpenTrade = nil
		pos.openTradeMutex.Unlock()

		pos.waitingToCloseTradeMutex.Lock()
		pos.WaitingToCloseTrade = false
		pos.waitingToCloseTradeMutex.Unlock()
	}

	if trade == nil {
		return nil, fmt.Errorf("something went wrong")
	}

	return trade, nil
}

func (pos *MarketOpenStrategy) GetOpenTrade() *MarketOpenTrade {
	pos.openTradeMutex.RLock()
	defer pos.openTradeMutex.RUnlock()
	if pos.OpenTrade == nil {
		return nil
	}
	copyTrade := &MarketOpenTrade{}
	copier.CopyWithOption(copyTrade, pos.OpenTrade, copier.Option{DeepCopy: true})
	return copyTrade
}

func (pos *MarketOpenStrategy) CreateOnCandleForBacktest() func(*fyersTypes.FyersHistoricalCandle) {
	backtestOnCandle := func(candle *fyersTypes.FyersHistoricalCandle) {
		pos.OnCandle(candle)
		/* Exit First, then Enter, because both can happen on the same candle */
		// pos.waitingToCloseTradeMutex.RLock()
		// if pos.WaitingToCloseTrade {
		// 	pos.waitingToCloseTradeMutex.RUnlock()
		// 	pos.Exit(candle.Close, false, time.Unix(candle.TS, 0))
		// } else {
		// 	pos.waitingToCloseTradeMutex.RUnlock()
		// }
	}
	return backtestOnCandle
}

func (pos *MarketOpenStrategy) CreateOnTickForBacktest(calculateLots func(candle *fyersTypes.FyersHistoricalCandle) int) func(*fyersTypes.FyersHistoricalCandle) {
	backtestOnTick := func(candle *fyersTypes.FyersHistoricalCandle) {
		pos.OnTick(candle)

		/* Exit First, then Enter, because both can happen on the same candle */
		pos.waitingToCloseTradeMutex.RLock()
		if pos.WaitingToCloseTrade {
			pos.waitingToCloseTradeMutex.RUnlock()
			pos.Exit(candle.Close, false, time.Unix(candle.TS, 0))
		} else {
			pos.waitingToCloseTradeMutex.RUnlock()
		}

		pos.waitingToOpenTradeMutex.RLock()
		if pos.WaitingToOpenTrade {
			pos.waitingToOpenTradeMutex.RUnlock()
			pos.Enter(candle.Close, calculateLots(candle), time.Unix(candle.TS, 0), marketConstants.TradeTypeNone)
		} else {
			pos.waitingToOpenTradeMutex.RUnlock()
		}
	}
	return backtestOnTick
}

func (pos *MarketOpenStrategy) OnWarmUpComplete() {
	pos.waitingToCloseTradeMutex.RLock()
	pos.lastCandleMutex.RLock()
	pos.openTradeMutex.RLock()
	if pos.WaitingToCloseTrade && pos.onCanExit != nil && pos.LastCandle != nil {
		pos.onCanExit(pos.ID, pos.TimeFrame, pos.Instrument, pos.UserID, pos.getExitReason(pos.LastCandle, pos.OpenTrade), pos.LastCandle.Candle.Close, pos.OpenTrade.PL)
	}
	pos.openTradeMutex.RUnlock()
	pos.lastCandleMutex.RUnlock()
	pos.waitingToCloseTradeMutex.RUnlock()

	pos.waitingToOpenTradeMutex.RLock()
	pos.lastCandleMutex.RLock()
	pos.firstCandleOfTheDayMutex.RLock()
	if pos.WaitingToOpenTrade && pos.onCanEnter != nil && pos.LastCandle != nil {
		pos.onCanEnter(pos.ID, pos.TimeFrame, pos.Instrument, pos.UserID, pos.getTradeType(pos.LastCandle, pos.FirstCandleOfTheDay), pos.LastCandle.Candle.Close)
	}
	pos.firstCandleOfTheDayMutex.RUnlock()
	pos.lastCandleMutex.RUnlock()
	pos.waitingToOpenTradeMutex.RUnlock()
}

func BacktestMarketOpenStrategy(backtestSDK *backtestSDK.BacktestSDK, instrument string, fromDate string, toDate string, lots int) (float64, float64, float64, int, int, int, float64, []*MarketOpenTrade) {
	pos := NewMarketOpenStrategy(instrument)
	pos.SetIsBacktestModeEnabled(true)
	backtestSDK.SubscribeCandle(instrument, marketOpenStrategyConstants.TimeFrame, pos.CreateOnCandleForBacktest())
	backtestSDK.SubscribeTick(instrument, marketOpenStrategyConstants.TimeFrame, pos.CreateOnTickForBacktest(func(candle *fyersTypes.FyersHistoricalCandle) int {
		return lots
	}))

	start, _ := time.Parse(time.RFC3339, fromDate)
	end, _ := time.Parse(time.RFC3339, toDate)
	backtestSDK.Backtest(instrument, marketOpenStrategyConstants.TimeFrame, start, end)

	loss := 0.
	lossCount := 0
	profit := 0.
	profitCount := 0
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

	return profit, loss, brokerage, profitCount, lossCount, profitCount + lossCount, finalPL, pos.ClosedTrades
}
