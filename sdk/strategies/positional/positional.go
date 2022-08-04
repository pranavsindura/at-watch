package strategies

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	positionalStrategyConstants "github.com/pranavsindura/at-watch/constants/strategies/positional"
	backtestSDK "github.com/pranavsindura/at-watch/sdk/backtest"
	indicatorsSDK "github.com/pranavsindura/at-watch/sdk/indicators"
	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	indicatorTypes "github.com/pranavsindura/at-watch/types/indicators"
	"github.com/pranavsindura/at-watch/utils"
	marketUtils "github.com/pranavsindura/at-watch/utils/market"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PositionalConfig struct {
	SuperTrend indicatorTypes.SuperTrendConfig
}

type PositionalIndicators struct {
	SuperTrend *indicatorTypes.SuperTrendData `json:"superTrend"`
}

type PositionalCandle struct {
	Candle     *fyersTypes.FyersHistoricalCandle `json:"candle"`
	Indicators *PositionalIndicators             `json:"indicators"`
}

type PositionalTrade struct {
	ID             primitive.ObjectID `json:"id"`
	TradeType      int                `json:"tradeType"`
	TradeTypeText  string             `json:"tradeTypeText"`
	Lots           int                `json:"lots"`
	Entry          *PositionalCandle  `json:"entry"`
	PL             float64            `json:"PL"`
	Exit           *PositionalCandle  `json:"exit"`
	ExitReason     int                `json:"exitReason"`
	ExitReasonText string             `json:"exitReasonText"`
	Brokerage      float64            `json:"brokerage"`
	UpdatedAtTS    int64              `json:"updatedAtTS"`
}

type PositionalStrategy struct {
	UserID                   primitive.ObjectID `json:"userID"`
	ID                       primitive.ObjectID `json:"id"`
	Instrument               string             `json:"instrument"`
	TimeFrame                int                `json:"timeFrame"`
	OpenTrade                *PositionalTrade   `json:"openTrade"`
	ClosedTrades             []*PositionalTrade `json:"closedTrades"`
	WaitingToOpenTrade       bool               `json:"waitingToOpenTrade"`
	WaitingToCloseTrade      bool               `json:"waitingToCloseTrade"`
	LastCandle               *PositionalCandle  `json:"lastCandle"`
	Config                   PositionalConfig   `json:"config"`
	openTradeMutex           *sync.RWMutex
	closedTradesMutex        *sync.RWMutex
	waitingToOpenTradeMutex  *sync.RWMutex
	waitingToCloseTradeMutex *sync.RWMutex
	lastCandleMutex          *sync.RWMutex
	IsBacktestModeEnabled    bool
	onCanEnter               func(ID primitive.ObjectID, userID primitive.ObjectID, tradeType int, candleClose float64)
	onCanExit                func(ID primitive.ObjectID, userID primitive.ObjectID, exitReason int, candleClose float64, PL float64)
}

func NewPositionalStrategy(instrument string, timeFrame int) *PositionalStrategy {
	strategy := &PositionalStrategy{
		Instrument:          instrument,
		TimeFrame:           timeFrame,
		OpenTrade:           nil,
		ClosedTrades:        make([]*PositionalTrade, 0),
		WaitingToOpenTrade:  false,
		WaitingToCloseTrade: false,
		LastCandle:          nil,
		Config: PositionalConfig{
			SuperTrend: indicatorTypes.SuperTrendConfig{
				Period:     positionalStrategyConstants.SuperTrendPeriod,
				Multiplier: positionalStrategyConstants.SuperTrendMultiplier,
			},
		},
		openTradeMutex:           &sync.RWMutex{},
		closedTradesMutex:        &sync.RWMutex{},
		waitingToOpenTradeMutex:  &sync.RWMutex{},
		waitingToCloseTradeMutex: &sync.RWMutex{},
		lastCandleMutex:          &sync.RWMutex{},
		IsBacktestModeEnabled:    false,
		onCanEnter:               nil,
		onCanExit:                nil,
	}
	return strategy
}

func (pos *PositionalStrategy) generateTradeId() primitive.ObjectID {
	return primitive.NewObjectID()
}

func (pos *PositionalStrategy) SetIsBacktestModeEnabled(enabled bool) {
	pos.IsBacktestModeEnabled = enabled
}

func (pos *PositionalStrategy) SetUserID(userID primitive.ObjectID) {
	pos.UserID = userID
}

func (pos *PositionalStrategy) SetID(ID primitive.ObjectID) {
	pos.ID = ID
}

func (pos *PositionalStrategy) SetOnCanEnter(onCanEnter func(ID primitive.ObjectID, userID primitive.ObjectID, tradeType int, candleClose float64)) {
	pos.onCanEnter = onCanEnter
}

func (pos *PositionalStrategy) SetOnCanExit(onCanExit func(ID primitive.ObjectID, userID primitive.ObjectID, exitReason int, candleClose float64, PL float64)) {
	pos.onCanExit = onCanExit
}

func (pos *PositionalStrategy) getTradeType(candle *PositionalCandle) int {
	if !candle.Indicators.SuperTrend.IsUsable {
		return marketConstants.TradeTypeNone
	}

	if candle.Indicators.SuperTrend.SuperTrendDirection {
		return marketConstants.TradeTypeBuy
	} else {
		return marketConstants.TradeTypeSell
	}
}

func (pos *PositionalStrategy) canEnter(candle *PositionalCandle) bool {
	return pos.getTradeType(candle) != marketConstants.TradeTypeNone
}

func (pos *PositionalStrategy) enter(candle *PositionalCandle, lots int, forceTradeType int) *PositionalTrade {
	copyCandle := &PositionalCandle{}
	copier.Copy(copyCandle, candle)
	var trade *PositionalTrade = nil
	var tradeType int
	if forceTradeType != marketConstants.TradeTypeNone {
		tradeType = forceTradeType
	} else {
		tradeType = pos.getTradeType(copyCandle)
	}

	switch tradeType {
	case marketConstants.TradeTypeBuy:
		trade = &PositionalTrade{
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
		}
	case marketConstants.TradeTypeSell:
		trade = &PositionalTrade{
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
		}
	}

	return trade
}

func (pos *PositionalStrategy) getExitReason(candle *PositionalCandle, trade *PositionalTrade) int {
	switch trade.TradeType {
	case marketConstants.TradeTypeBuy:
		if pos.getTradeType(candle) == marketConstants.TradeTypeSell {
			return marketConstants.TradeExitReasonSystemForceExit
		} else {
			return marketConstants.TradeExitReasonNone
		}
	case marketConstants.TradeTypeSell:
		if pos.getTradeType(candle) == marketConstants.TradeTypeBuy {
			return marketConstants.TradeExitReasonSystemForceExit
		} else {
			return marketConstants.TradeExitReasonNone
		}
	default:
		return marketConstants.TradeExitReasonNone
	}
}

func (pos *PositionalStrategy) canExit(candle *PositionalCandle, trade *PositionalTrade) bool {
	return pos.getExitReason(candle, trade) != marketConstants.TradeExitReasonNone
}

func (pos *PositionalStrategy) exit(candle *PositionalCandle, trade *PositionalTrade, forceExit bool) *PositionalTrade {
	copyCandle := &PositionalCandle{}
	copier.Copy(copyCandle, candle)

	var exitReason int
	if forceExit {
		exitReason = marketConstants.TradeExitReasonUserForceExit
	} else {
		exitReason = pos.getExitReason(candle, trade)
	}

	var completedTrade *PositionalTrade = nil

	switch trade.TradeType {
	case marketConstants.TradeTypeBuy:
		PL := (candle.Candle.Close - trade.Entry.Candle.Close) * float64(trade.Lots)

		completedTrade = &PositionalTrade{}
		copier.Copy(completedTrade, trade)
		completedTrade.PL = PL
		completedTrade.Exit = copyCandle
		completedTrade.ExitReason = exitReason
		completedTrade.ExitReasonText = marketConstants.TradeExitReasonToTextMap[exitReason]
		completedTrade.Brokerage = marketUtils.CalculateZerodhaNSEFuturesCharges(trade.Entry.Candle.Close, candle.Candle.Close, trade.Lots)
	case marketConstants.TradeTypeSell:
		PL := -(candle.Candle.Close - trade.Entry.Candle.Close) * float64(trade.Lots)

		completedTrade = &PositionalTrade{}
		copier.Copy(completedTrade, trade)
		completedTrade.PL = PL
		completedTrade.Exit = copyCandle
		completedTrade.ExitReason = exitReason
		completedTrade.ExitReasonText = marketConstants.TradeExitReasonToTextMap[exitReason]
		completedTrade.Brokerage = marketUtils.CalculateZerodhaNSEFuturesCharges(trade.Entry.Candle.Close, candle.Candle.Close, trade.Lots)
	}

	return completedTrade
}

func (pos *PositionalStrategy) update(candle *PositionalCandle, trade *PositionalTrade) *PositionalTrade {
	var updatedTrade *PositionalTrade = &PositionalTrade{}
	copier.Copy(updatedTrade, trade)

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

func (pos *PositionalStrategy) onCandle(candleData *PositionalCandle) {
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
		fmt.Println("waiting to exit trade", marketConstants.TradeExitReasonToTextMap[pos.getExitReason(candleData, pos.OpenTrade)])
		pos.waitingToCloseTradeMutex.Lock()
		pos.WaitingToCloseTrade = true
		pos.waitingToCloseTradeMutex.Unlock()

		if pos.onCanExit != nil {
			pos.onCanExit(pos.ID, pos.UserID, pos.getExitReason(candleData, pos.OpenTrade), candleData.Candle.Close, pos.OpenTrade.PL)
		}
	}

	if pos.canEnter(candleData) && (pos.OpenTrade == nil || pos.OpenTrade.TradeType != pos.getTradeType(candleData)) {
		fmt.Println("waiting to enter trade, but must exit existing trade if any", marketConstants.TradeTypeToTextMap[pos.getTradeType(candleData)])
		pos.waitingToOpenTradeMutex.Lock()
		pos.WaitingToOpenTrade = true
		pos.waitingToOpenTradeMutex.Unlock()
		if pos.onCanEnter != nil {
			pos.onCanEnter(pos.ID, pos.UserID, pos.getTradeType(candleData), candleData.Candle.Close)
		}
	}
	pos.openTradeMutex.RUnlock()
}

func (pos *PositionalStrategy) createCandle(candle *fyersTypes.FyersHistoricalCandle) *PositionalCandle {
	var lastCandle *fyersTypes.FyersHistoricalCandle = nil
	var lastSuperTrendData *indicatorTypes.SuperTrendData = nil
	pos.lastCandleMutex.RLock()
	if pos.LastCandle != nil {
		lastCandle = &fyersTypes.FyersHistoricalCandle{}
		copier.Copy(lastCandle, pos.LastCandle.Candle)
		lastSuperTrendData = &indicatorTypes.SuperTrendData{}
		copier.Copy(lastSuperTrendData, pos.LastCandle.Indicators.SuperTrend)
	}
	pos.lastCandleMutex.RUnlock()
	currentSuperTrendData := indicatorsSDK.GetSuperTrend(lastCandle, lastSuperTrendData, candle, pos.Config.SuperTrend)
	var copyCandle *fyersTypes.FyersHistoricalCandle = &fyersTypes.FyersHistoricalCandle{}
	copier.Copy(copyCandle, candle)
	posCandle := &PositionalCandle{
		Candle: copyCandle,
		Indicators: &PositionalIndicators{
			SuperTrend: currentSuperTrendData,
		},
	}

	return posCandle
}

func (pos *PositionalStrategy) OnCandle(candle *fyersTypes.FyersHistoricalCandle) {
	posCandle := pos.createCandle(candle)
	pos.lastCandleMutex.Lock()
	pos.LastCandle = &PositionalCandle{}
	copier.Copy(pos.LastCandle, posCandle)
	pos.lastCandleMutex.Unlock()
	pos.onCandle(posCandle)
}

func (pos *PositionalStrategy) OnTick(candle *fyersTypes.FyersHistoricalCandle) {
	// we will fabricate a candle and update the open trade based on it directly
	// based on whatever the pos.update func needs from the candleData
	candleData := &PositionalCandle{
		Candle: &fyersTypes.FyersHistoricalCandle{
			Close: candle.Close,
		},
	}
	pos.openTradeMutex.Lock()
	if pos.OpenTrade != nil {
		pos.OpenTrade = pos.update(candleData, pos.OpenTrade)
		pos.OpenTrade.UpdatedAtTS = candle.TS
	}
	pos.openTradeMutex.Unlock()
}

func (pos *PositionalStrategy) Enter(price float64, lots int, entryAt time.Time, forceTradeType int) (*PositionalTrade, error) {
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

	tradeType := pos.getTradeType(posCandle)
	if !pos.IsBacktestModeEnabled {
		if tradeType == marketConstants.TradeTypeNone {
			return nil, fmt.Errorf("trade type is " + marketConstants.TradeTypeNoneText + ", unable to enter")
		}
		if tradeType != forceTradeType {
			return nil, fmt.Errorf("trade type is " + marketConstants.TradeTypeToTextMap[tradeType] + ", but received " + marketConstants.TradeTypeToTextMap[forceTradeType])
		}
	}

	trade := pos.enter(posCandle, lots, forceTradeType)
	if trade != nil {
		pos.openTradeMutex.Lock()
		pos.OpenTrade = &PositionalTrade{}
		copier.Copy(pos.OpenTrade, trade)
		pos.openTradeMutex.Unlock()

		pos.waitingToOpenTradeMutex.Lock()
		pos.WaitingToOpenTrade = false
		pos.waitingToOpenTradeMutex.Unlock()
	}

	if trade == nil {
		return nil, fmt.Errorf("something went wrong")
	}

	return trade, nil
}

func (pos *PositionalStrategy) Exit(price float64, forceExit bool, exitAt time.Time) (*PositionalTrade, error) {
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

func (pos *PositionalStrategy) GetOpenTrade() *PositionalTrade {
	pos.openTradeMutex.RLock()
	defer pos.openTradeMutex.RUnlock()
	if pos.OpenTrade == nil {
		return nil
	}
	copyTrade := &PositionalTrade{}
	copier.Copy(copyTrade, pos.OpenTrade)
	return copyTrade
}

func (pos *PositionalStrategy) CreateOnCandleForBacktest(calculateLots func(candle *fyersTypes.FyersHistoricalCandle) int) func(*fyersTypes.FyersHistoricalCandle) {
	backtestOnCandle := func(candle *fyersTypes.FyersHistoricalCandle) {
		pos.OnCandle(candle)

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
	return backtestOnCandle
}

func (pos *PositionalStrategy) OnWarmUpComplete() {
	pos.waitingToCloseTradeMutex.RLock()
	pos.lastCandleMutex.RLock()
	pos.openTradeMutex.RLock()
	if pos.WaitingToCloseTrade && pos.onCanExit != nil && pos.LastCandle != nil {
		pos.onCanExit(pos.ID, pos.UserID, pos.getExitReason(pos.LastCandle, pos.OpenTrade), pos.LastCandle.Candle.Close, pos.OpenTrade.PL)
	}
	pos.openTradeMutex.RUnlock()
	pos.lastCandleMutex.RUnlock()
	pos.waitingToCloseTradeMutex.RUnlock()

	pos.waitingToOpenTradeMutex.RLock()
	pos.lastCandleMutex.RLock()
	if pos.WaitingToOpenTrade && pos.onCanEnter != nil && pos.LastCandle != nil {
		pos.onCanEnter(pos.ID, pos.UserID, pos.getTradeType(pos.LastCandle), pos.LastCandle.Candle.Close)
	}
	pos.lastCandleMutex.RUnlock()
	pos.waitingToOpenTradeMutex.RUnlock()
}

func BacktestPositionalStrategy(backtestSDK *backtestSDK.BacktestSDK, instrument string, timeFrame int, fromDate string, toDate string, lots int) (float64, float64, float64, int, int, int, float64, []*PositionalTrade) {
	pos := NewPositionalStrategy(instrument, timeFrame)
	pos.SetIsBacktestModeEnabled(true)
	backtestSDK.SubscribeCandle(instrument, marketConstants.TimeFrame15m, pos.CreateOnCandleForBacktest(func(candle *fyersTypes.FyersHistoricalCandle) int {
		return lots
	}))

	start, _ := time.Parse(time.RFC3339, fromDate)
	end, _ := time.Parse(time.RFC3339, toDate)
	backtestSDK.Backtest(instrument, timeFrame, start, end)

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
