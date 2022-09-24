package fyersSDK

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/chromedp/chromedp"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
	fyersConstants "github.com/pranavsindura/at-watch/constants/fyers"
	fyersWatch "github.com/pranavsindura/at-watch/sdk/fyersWatch"
	fyersWatchAPI "github.com/pranavsindura/at-watch/sdk/fyersWatch/api"
	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	fyersUtils "github.com/pranavsindura/at-watch/utils/fyers"
	"github.com/rs/zerolog/log"
)

var fyersAccessTokenMutex *sync.RWMutex = &sync.RWMutex{}
var fyersAccessToken string

// var shouldStopMarketWatch = make(chan struct{}, 1)
var isMarketWatchActiveMutex *sync.RWMutex = &sync.RWMutex{}
var isMarketWatchActive = false
var cli *fyersWatch.WatchNotifier = nil

// var marketWatchEventEmitter = eventemitter.NewEmitter(true)

func getAppSecretHash() string {
	fyersSecretID := getFyersSecretID()
	fyersAppID := getFyersAppID()
	appSecret := fyersAppID + ":" + fyersSecretID
	appSecretHash := sha256.New()
	appSecretHash.Write([]byte(appSecret))
	appSecretHashSum := hex.EncodeToString(appSecretHash.Sum(nil))
	return appSecretHashSum
}

func getFyersAppID() string {
	fyersAppID := os.Getenv(envConstants.FyersAppID)
	return fyersAppID
}

func getFyersSecretID() string {
	fyersSecretID := os.Getenv(envConstants.FyersSecretID)
	return fyersSecretID
}

func getFyersRedirectURL() string {
	fyersRedirectURL := os.Getenv(envConstants.FyersRedirectURL)
	return fyersRedirectURL
}

func GetFyersAccessToken() string {
	fyersAccessTokenMutex.RLock()
	token := fyersAccessToken
	fyersAccessTokenMutex.RUnlock()
	return token
}

func SetFyersAccessToken(newFyersAccessToken string) {
	fyersAccessTokenMutex.Lock()
	fyersAccessToken = newFyersAccessToken
	fyersAccessTokenMutex.Unlock()
}

func getAuthorizationHeader() string {
	fyersAppID := getFyersAppID()
	return fyersAppID + ":" + GetFyersAccessToken()
}

func GenerateAuthCodeURL() string {
	fyersAppID := getFyersAppID()
	fyersRedirectURL := getFyersRedirectURL()
	url := fyersConstants.EndpointAPI + fyersConstants.GenerateAuthCodeURL
	url = strings.Replace(url, "##CLIENT_ID##", fyersAppID, 1)
	url = strings.Replace(url, "##REDIRECT_URI##", fyersRedirectURL, 1)
	return url
}

func generateValidateAuthCodeUrl() string {
	return fyersConstants.EndpointAPI + fyersConstants.ValidateAuthCodeURL
}

func ValidateAuthCode(authCode string) (string, error) {
	appSecretHash := getAppSecretHash()
	data := map[string]string{
		"grant_type": "authorization_code",
		"appIdHash":  appSecretHash,
		"code":       authCode,
	}

	dataAsJSON, err := json.Marshal(data)

	if err != nil {
		log.Fatal().Err(err)
	}

	res, err := http.Post(generateValidateAuthCodeUrl(), "application/json", bytes.NewBuffer(dataAsJSON))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var body fyersTypes.ValidateAuthCodeResponse
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf(body.Message)
	}

	return body.AccessToken, nil
}

func generateFetchHistoricalDataURL() string {
	return fyersConstants.EndpointDataRest + fyersConstants.HistoricalDataURL
}

func FetchHistoricalData(symbol string, resolution string, rangeFrom int64, rangeTo int64, contFlag int) ([]fyersTypes.FyersHistoricalCandle, error) {
	onError := func(err error) ([]fyersTypes.FyersHistoricalCandle, error) {
		return make([]fyersTypes.FyersHistoricalCandle, 0), err
	}

	req, err := http.NewRequest(http.MethodGet, generateFetchHistoricalDataURL(), nil)
	if err != nil {
		return onError(err)
	}

	query := req.URL.Query()
	query.Add("symbol", fyersConstants.Exchange+":"+symbol)
	query.Add("resolution", resolution)
	query.Add("date_format", "0")
	query.Add("range_from", strconv.FormatInt(rangeFrom, 10))
	query.Add("range_to", strconv.FormatInt(rangeTo, 10))
	query.Add("cont_flag", strconv.Itoa(contFlag))

	req.URL.RawQuery = query.Encode()

	req.Header.Add("Authorization", getAuthorizationHeader())

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return onError(err)
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return onError(err)
	}

	var body fyersTypes.FyersHistoricalDataResponse
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return onError(err)
	}

	transformedData, err := fyersUtils.TransformHistoricalData(body)

	return transformedData, err
}

func generateNSECapitalMarketSymbolsURL() string {
	return fyersConstants.EndpointPublic + fyersConstants.NSECapitalMarketSymbolsURL
}

func IsValidInstrument(instrument string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, generateNSECapitalMarketSymbolsURL(), nil)
	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()
	csvReader := csv.NewReader(res.Body)
	records, err := csvReader.ReadAll()

	if err != nil {
		return false, err
	}

	for _, record := range records {
		if len(record) != len(fyersConstants.NSECapitalMarketSymbolsHeaders) {
			return false, fmt.Errorf("IsValidInstrument - symbol headers length do not match the expected length")
		}
		// 10th element is the Symbol Ticker acc to fyersConstants.NSECapitalMarketSymbolsHeaders
		symbol := ""
		if len(record[9]) <= 4 {
			return false, fmt.Errorf("IsValidInstrument - symbol length <= 4, cannot remove exchange \"NSE:\" from it - " + record[9])
		}
		symbol = record[9][4:]
		if symbol == instrument {
			return true, nil
		}
	}

	return false, nil
}

func IsValidInstrumentMany(instruments []string) (map[string]bool, error) {
	req, err := http.NewRequest(http.MethodGet, generateNSECapitalMarketSymbolsURL(), nil)
	if err != nil {
		return make(map[string]bool), err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return make(map[string]bool), err
	}

	defer res.Body.Close()
	csvReader := csv.NewReader(res.Body)
	records, err := csvReader.ReadAll()

	if err != nil {
		return make(map[string]bool), err
	}

	realInstrumentsMap := make(map[string]bool)

	for _, record := range records {
		if len(record) != len(fyersConstants.NSECapitalMarketSymbolsHeaders) {
			return make(map[string]bool), fmt.Errorf("IsValidInstrumentMany - symbol headers length do not match the expected length")
		}
		// 10th element is the Symbol Ticker acc to fyersConstants.NSECapitalMarketSymbolsHeaders
		symbol := ""
		if len(record[9]) <= 4 {
			return make(map[string]bool), fmt.Errorf("IsValidInstrumentMany - symbol length <= 4, cannot remove exchange \"NSE:\" from it - " + record[9])
		}
		symbol = record[9][4:]
		realInstrumentsMap[symbol] = true
	}

	validInstrumentsMap := make(map[string]bool)
	for _, instrument := range instruments {
		_, ok := realInstrumentsMap[instrument]
		validInstrumentsMap[instrument] = ok
	}

	return validInstrumentsMap, nil
}

func GetValidInstruments() ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, generateNSECapitalMarketSymbolsURL(), nil)
	if err != nil {
		return make([]string, 0), err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return make([]string, 0), err
	}

	defer res.Body.Close()
	csvReader := csv.NewReader(res.Body)
	records, err := csvReader.ReadAll()

	if err != nil {
		return make([]string, 0), err
	}

	instruments := make([]string, 0)

	for _, record := range records {
		if len(record) != len(fyersConstants.NSECapitalMarketSymbolsHeaders) {
			return make([]string, 0), fmt.Errorf("GetValidInstruments - symbol headers length do not match the expected length")
		}
		// 10th element is the Symbol Ticker acc to fyersConstants.NSECapitalMarketSymbolsHeaders
		symbol := ""
		if len(record[9]) <= 4 {
			return make([]string, 0), fmt.Errorf("GetValidInstruments - symbol length <= 4, cannot remove exchange \"NSE:\" from it - " + record[9])
		}
		symbol = record[9][4:]
		fmt.Println(symbol)
		instruments = append(instruments, symbol)
	}

	return instruments, nil
}

func setIsMarketWatchActive(value bool) {
	isMarketWatchActiveMutex.Lock()
	isMarketWatchActive = value
	isMarketWatchActiveMutex.Unlock()
}

func StartMarketWatch(instruments []string, onMarketWatchConnect func(), onMarketWatchMessage func(fyersWatchAPI.Notification), onMarketWatchError func(error), onMarketWatchDisconnect func(error)) (bool, error) {
	if IsMarketWatchActive() {
		return false, fmt.Errorf("market watch is already started")
	}
	instrumentsWithExchangePrefix := make([]string, 0)
	for _, instrument := range instruments {
		instrumentsWithExchangePrefix = append(instrumentsWithExchangePrefix, fyersConstants.Exchange+":"+instrument)
	}

	apiKey := getFyersAppID()
	accessToken := GetFyersAccessToken()
	cli = fyersWatch.NewNotifier(apiKey, accessToken).
		WithOnConnectFunc(onMarketWatchConnect).
		WithOnMessageFunc(onMarketWatchMessage).
		WithOnErrorFunc(onMarketWatchError).
		WithOnDisconnectFunc(onMarketWatchDisconnect)

	fmt.Println(instrumentsWithExchangePrefix)

	cli.Subscribe(fyersWatchAPI.SymbolDataTick, instrumentsWithExchangePrefix...)
	setIsMarketWatchActive(true)

	return true, nil
}

func IsMarketWatchActive() bool {
	isMarketWatchActiveMutex.RLock()
	isActive := isMarketWatchActive
	isMarketWatchActiveMutex.RUnlock()
	return isActive
}

func StopMarketWatch() (bool, error) {
	if IsMarketWatchActive() {
		cli.Disconnect()
		setIsMarketWatchActive(false)
		return true, nil
	} else {
		return false, fmt.Errorf("market watch is already stopped")
	}
}

func AutomateAdminLogin() (bool, error) {
	browserCtx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	loginURL := GenerateAuthCodeURL()

	dispatchKeyboardEventJS := func(qs string, c rune) string {
		return `document.querySelector("` + qs + `").dispatchEvent(new KeyboardEvent("keydown", {
			key: "` + string(c) + `",
			keyCode: ` + strconv.Itoa(int(c)) + `,
			code: "Key` + string(c) + `",
		}));`
	}

	err := chromedp.Run(
		browserCtx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible("#fy_client_id", chromedp.ByID),
		chromedp.SendKeys("#fy_client_id", os.Getenv(envConstants.AdminClientID), chromedp.ByID),
		chromedp.WaitVisible("#clientIdSubmit", chromedp.ByID),
		chromedp.Click("#clientIdSubmit", chromedp.ByID),
		chromedp.WaitVisible("#fy_client_pwd", chromedp.ByID),
		chromedp.SendKeys("#fy_client_pwd", os.Getenv(envConstants.AdminClientPassword), chromedp.ByID),
		chromedp.WaitVisible("#loginSubmit", chromedp.ByID),
		chromedp.Click("#loginSubmit", chromedp.ByID),
		chromedp.WaitVisible("#pin-container", chromedp.ByID),
		chromedp.Evaluate(dispatchKeyboardEventJS("#pin-container > #first", rune(os.Getenv(envConstants.AdminClientPin)[0])), nil),
		chromedp.Evaluate(dispatchKeyboardEventJS("#pin-container > #second", rune(os.Getenv(envConstants.AdminClientPin)[1])), nil),
		chromedp.Evaluate(dispatchKeyboardEventJS("#pin-container > #third", rune(os.Getenv(envConstants.AdminClientPin)[2])), nil),
		chromedp.Evaluate(dispatchKeyboardEventJS("#pin-container > #fourth", rune(os.Getenv(envConstants.AdminClientPin)[3])), nil),
		chromedp.WaitVisible("#verifyPinSubmit", chromedp.ByID),
		chromedp.Click("#verifyPinSubmit", chromedp.ByID),
		chromedp.WaitVisible("#done", chromedp.ByID),
	)

	if err != nil {
		return false, err
	}

	return true, nil
}
