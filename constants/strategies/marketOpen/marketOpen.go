package marketOpenStrategyConstants

import (
	"time"

	marketConstants "github.com/pranavsindura/at-watch/constants/market"
)

const TimeFrame int = marketConstants.TimeFrame15m
const TimeFrameDuration time.Duration = time.Minute * 15
const CheckCandleCount int = 3 // check 3 candles after first candle - 2,3,4
// could be based on expected daily volatility
const TargetBuffer float64 = 100.  // 100pt
const StopLossBuffer float64 = 50. // 50pt
const CutOffHour = 14              // 2:30pm
const CutOffMinute = 30
