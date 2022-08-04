package marketTypes

type MarketTick struct {
	LTP    float64 `json:"ltp"`
	TS     int64   `json:"ts"`
	Volume float64 `json:"volume"`
}
