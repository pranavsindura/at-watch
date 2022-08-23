package marketUtils

import (
	"math"
	"time"

	"github.com/gorhill/cronexpr"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
)

func GetCandleTimeOf(timeFrame int, fromDate time.Time) time.Time {
	switch timeFrame {
	case marketConstants.TimeFrame15m:
		nowMinus15m := fromDate.Add(-time.Minute * 15)
		currentTickCandleTimestamp := cronexpr.MustParse(marketConstants.Cron15m).Next(nowMinus15m)
		return currentTickCandleTimestamp
	}
	return time.Now()
}

func CalculateZerodhaFuturesBrokerage(price float64, qty float64) float64 {
	return math.Min(price*qty*marketConstants.ZerodhaFuturesBrokeragePerc, marketConstants.ZerodhaFuturesBrokerageFixed)
}

func CalculateZerodhaNSEFuturesCharges(buyPrice float64, sellPrice float64, qty int) float64 {
	turnover := (buyPrice + sellPrice) * float64(qty)
	zerodhaBrokerage := CalculateZerodhaFuturesBrokerage(buyPrice, float64(qty)) + CalculateZerodhaFuturesBrokerage(sellPrice, float64(qty))
	stt := sellPrice * float64(qty) * marketConstants.FuturesSTTPerc
	nseTxnCharge := turnover * marketConstants.NSETXNChargePerc
	gst := (zerodhaBrokerage + nseTxnCharge) * marketConstants.GSTPerc
	sebiCharge := turnover * marketConstants.SEBICharge
	stampDuty := buyPrice * float64(qty) * marketConstants.StampDuty

	charges := math.Ceil(zerodhaBrokerage + stt + nseTxnCharge + gst + sebiCharge + stampDuty)
	return charges
}
