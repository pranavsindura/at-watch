package fyersConstants

const Exchange string = "NSE"
const EndpointAPI string = "https://api.fyers.in/api/v2"
const EndpointDataRest string = "https://api.fyers.in/data-rest/v2"
const EndpointPublic string = "https://public.fyers.in"
const GenerateAuthCodeURL string = "/generate-authcode?client_id=##CLIENT_ID##&redirect_uri=##REDIRECT_URI##&response_type=code&state=sample_state"
const ValidateAuthCodeURL string = "/validate-authcode"
const HistoricalDataURL string = "/history"
const NSECapitalMarketSymbolsURL string = "/sym_details/NSE_CM.csv"

const (
	FyersEventTick string = "FYERS_EVENT_TICK"
)

var NSECapitalMarketSymbolsHeaders = []string{
	"FyToken",
	"Symbol Details",
	"Exchange Instrument type",
	"Minimum lot size",
	"Tick size",
	"ISIN",
	"Trading Session",
	"Last update date",
	"Expiry date",
	"Symbol ticker",
	"Exchange",
	"Segment",
	"Scrip code",
	"Underlying scrip code",
	"Symbol Name",
	"Strike price",
	"Option type",
}
