package fyersTypes

type FyersHistoricalDataResponse struct {
	Candles []([6]float64) `json:"candles"`
}

type FyersHistoricalCandle struct {
	TS         int64   `json:"ts"`
	DateString string  `json:"dateString"`
	Day        string  `json:"day"`
	Open       float64 `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     float64 `json:"volume"`
}

type ValidateAuthCodeResponse struct {
	Status      string `json:"s"`
	Code        int    `json:"code"`
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
}

type FyersInstrument struct {
	FyToken                string  `json:"fyToken"`
	SymbolDetails          string  `json:"symbolDetails"`
	ExchangeInstrumentType int32   `json:"exchangeInstrumentType"`
	MinimumLotSize         int32   `json:"minimumLotSize"`
	TickSize               float64 `json:"tickSize"`
	ISIN                   string  `json:"ISIN"`
	TradingSession         string  `json:"tradingSession"`
	LastUpdateDate         string  `json:"lastUpdateDate"`
	ExpiryDate             string  `json:"expiryDate"`
	SymbolTicker           string  `json:"symbolTicker"`
	Exchange               int32   `json:"exchange"`
	Segment                int32   `json:"segment"`
	ScripCode              int32   `json:"scripCode"`
	SymbolName             string  `json:"symbolName"`
	UnderlyingScripCode    int32   `json:"underlyingScripCode"`
	StrikePrice            float64 `json:"strikePrice"`
	OptionType             string  `json:"optionType"`
}
