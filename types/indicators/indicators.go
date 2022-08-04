package indicatorTypes

type SuperTrendData struct {
	Index               int       `json:"index"`
	ATR                 float64   `json:"atr"`
	PastTRList          []float64 `json:"pastTRList"`
	BasicUpperBound     float64   `json:"basicUpperBound"`
	BasicLowerBound     float64   `json:"basicLowerBound"`
	FinalUpperBound     float64   `json:"finalUpperBound"`
	FinalLowerBound     float64   `json:"finalLowerBound"`
	SuperTrend          float64   `json:"superTrend"`
	SuperTrendDirection bool      `json:"superTrendDirection"`
	IsUsable            bool      `json:"isUsable"`
}

type SuperTrendConfig struct {
	Period     int     `json:"period"`
	Multiplier float64 `json:"multiplier"`
}
