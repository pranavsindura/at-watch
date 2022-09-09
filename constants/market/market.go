// Type can have a None Type
// Type to Text Map can have a mapping from None to NoneText
// but a reverse map from Text to Type should not have NoneText to None mapping because None is only an internal type

package marketConstants

import "time"

const (
	MarketEventTick   string = "MARKET_EVENT_TICK"
	MarketEventCandle string = "MARKET_EVENT_CANDLE"
)

const (
	TimeFrameUnknown int = iota
	TimeFrame15m
)

const (
	TimeFrameUnknownText string = "UNKNOWN"
	TimeFrame15mText     string = "15m"
)

var TimeFrameToTextMap map[int]string = map[int]string{
	TimeFrameUnknown: TimeFrameUnknownText,
	TimeFrame15m:     TimeFrame15mText,
}

var TextToTimeFrameMap map[string]int = map[string]int{
	TimeFrame15mText: TimeFrame15m,
}

const (
	Cron15m = "*/15 * * * *"
)

const (
	TradeTypeNone int = iota
	TradeTypeBuy
	TradeTypeSell
)

const (
	TradeTypeNoneText string = "NONE"
	TradeTypeBuyText  string = "BUY"
	TradeTypeSellText string = "SELL"
)

var TradeTypeToTextMap = map[int]string{
	TradeTypeNone: TradeTypeNoneText,
	TradeTypeBuy:  TradeTypeBuyText,
	TradeTypeSell: TradeTypeSellText,
}

var TradeTypeTextToTypeMap = map[string]int{
	TradeTypeBuyText:  TradeTypeBuy,
	TradeTypeSellText: TradeTypeSell,
}

const (
	TradeExitReasonNone int = iota
	TradeExitReasonTargetReached
	TradeExitReasonStopLossHit
	TradeExitReasonSystemForceExit
	TradeExitReasonUserForceExit
)

const (
	TradeExitReasonNoneText            string = "NONE"
	TradeExitReasonTargetReachedText   string = "TARGET_REACHED"
	TradeExitReasonStopLossHitText     string = "STOP_LOSS_HIT"
	TradeExitReasonSystemForceExitText string = "SYSTEM_FORCE_EXIT"
	TradeExitReasonUserForceExitText   string = "USER_FORCE_EXIT"
)

const (
	TradeStatusUnknown int = iota
	TradeStatusOpen
	TradeStatusClosed
	// TradeStatusWaitingToOpen
	// TradeStatusWaitingToClose
)

const (
	TradeStatusUnknownText string = "UNKNOWN"
	TradeStatusOpenText    string = "OPEN"
	TradeStatusClosedText  string = "CLOSED"
	// TradeStatusWaitingToOpenText  string = "WAITING_TO_OPEN"
	// TradeStatusWaitingToCloseText string = "WAITING_TO_CLOSE"
)

var TradeExitReasonToTextMap map[int]string = map[int]string{
	TradeExitReasonNone:            TradeExitReasonNoneText,
	TradeExitReasonTargetReached:   TradeExitReasonTargetReachedText,
	TradeExitReasonStopLossHit:     TradeExitReasonStopLossHitText,
	TradeExitReasonSystemForceExit: TradeExitReasonSystemForceExitText,
	TradeExitReasonUserForceExit:   TradeExitReasonUserForceExitText,
}

const ZerodhaOptionsBrokerage float64 = 20.        // flat Rs. 20/-
const ZerodhaFuturesBrokeragePerc float64 = 0.0003 // 0.03% -> 0.0003
const ZerodhaFuturesBrokerageFixed float64 = 20
const OptionsSTTPerc float64 = 0.0005           // 0.05% on sell side (on premium)
const FuturesSTTPerc float64 = 0.0001           // 0.01% on sell side
const NSEFuturesTXNChargePerc float64 = 0.00002 // 0.002%
const NSEOptionsTXNChargePerc float64 = 0.00053 // 0.053%
const GSTPerc float64 = 0.18                    // 18%
const SEBICharge float64 = 0.000001             // Rs 10 per crore -> 10/10^7 -> 10^-6
const FuturesStampDuty float64 = 0.00002        // 0.002% or ₹200 / crore (on the buy side) -> 200/10^7 -> 2 * 10^-5
const OptionsStampDuty float64 = 0.00003        // 0.003% or ₹300 / crore (on the buy side) -> 300/10^7 -> 3 * 10^-5

const WarmUpDuration time.Duration = time.Duration(time.Hour * 24 * 365) // 1 Year
// For testing:
// const WarmUpDuration time.Duration = time.Duration(time.Hour * 24 * 90) // 3 Months

const (
	Resolution1m string = "1"
)

const (
	MarketOpenHours    int = 9
	MarketOpenMinutes  int = 15 // 0915 hours 9:15am
	MarketCloseHours   int = 15
	MarketCloseMinutes int = 30 // 1530 hours 3:30pm
)
