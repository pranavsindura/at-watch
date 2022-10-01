package marketOpenStrategyConstants

import (
	"time"

	marketConstants "github.com/pranavsindura/at-watch/constants/market"
)

const TimeFrame int = marketConstants.TimeFrame15m
const TimeFrameDuration time.Duration = time.Minute * 15 // TODO: Pranav - make this 5min
const CheckCandleCount int = 3                           // check 3 candles after first candle - 2,3,4 TODO: Pranav - 2nd breaks 1st or 3rd breaks 2nd
// could be based on expected daily volatility
const TargetBuffer float64 = 100.  // 100pt TODO: Pranav - ->75pt
const StopLossBuffer float64 = 50. // 50pt TODO: Pranav - ->15,20,25
const CutOffHour = 14              // 2:30pm
const CutOffMinute = 30
