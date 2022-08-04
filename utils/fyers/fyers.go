package fyersUtils

import (
	"fmt"

	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	"github.com/pranavsindura/at-watch/utils"
)

func TransformHistoricalData(data fyersTypes.FyersHistoricalDataResponse) ([]fyersTypes.FyersHistoricalCandle, error) {
	var transformedData []fyersTypes.FyersHistoricalCandle = make([]fyersTypes.FyersHistoricalCandle, 0)

	for _, candle := range data.Candles {
		if len(candle) != 6 {
			return make([]fyersTypes.FyersHistoricalCandle, 0), fmt.Errorf("candle length must be == 6")
		}
		timestampInSeconds := candle[0]
		open := candle[1]
		high := candle[2]
		low := candle[3]
		close := candle[4]
		volume := candle[5]

		transformedData = append(transformedData, fyersTypes.FyersHistoricalCandle{
			TS:         int64(timestampInSeconds),
			DateString: utils.GetDateStringFromTimestamp(int64(timestampInSeconds)),
			Day:        utils.GetWeekdayFromTimestamp(int64(timestampInSeconds)),
			Open:       open,
			High:       high,
			Low:        low,
			Close:      close,
			Volume:     volume,
		})
	}
	return transformedData, nil
}
