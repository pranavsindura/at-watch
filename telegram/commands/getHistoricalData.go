package telegramCommands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	telegramBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jinzhu/copier"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	"github.com/pranavsindura/at-watch/utils"
	telegramUtils "github.com/pranavsindura/at-watch/utils/telegram"
	"github.com/yukithm/json2csv"
)

func getHistoricalData(bot *telegramBot.BotAPI, update telegramBot.Update, instrument string, fromDateString string, toDateString string, resolution string) {
	bot.Send(telegramUtils.GenerateReplyMessage(update, "Fetching Historical Data"))

	fromDate, _ := time.Parse(time.RFC3339, fromDateString)
	toDate, _ := time.Parse(time.RFC3339, toDateString)

	currentDate := time.Time{}
	copier.CopyWithOption(&currentDate, &fromDate, copier.Option{DeepCopy: true})

	rows := make([]json2csv.KeyValue, 0)

	for currentDate.Sub(toDate).Milliseconds() <= 0 {
		after3Months := currentDate.AddDate(0, 3, 0)
		if toDate.Sub(after3Months).Milliseconds() < 0 {
			after3Months = toDate
		}

		data, err := fyersSDK.FetchHistoricalData(instrument, resolution, currentDate.Unix(), after3Months.Unix(), 0)
		if err != nil {
			fmt.Println("Error while fetching historical data", err)
			bot.Send(telegramUtils.GenerateReplyMessage(update, telegramUtils.GenerateGenericErrorText(fmt.Errorf("error while fetching historical data"))))
			return
		}

		dataJSON := utils.BruteStringify(data)
		var dataMapList []map[string]interface{}
		json.Unmarshal([]byte(dataJSON), &dataMapList)

		for _, tradesMap := range dataMapList {
			row, err := json2csv.JSON2CSV(tradesMap)
			if err != nil {
				fmt.Println("attempting to convert to csv failed", row, err)
			}
			if len(row) != 1 {
				fmt.Println("unexpected row length", row)
				continue
			}
			rows = append(rows, row[0])
		}

		currentDate = after3Months.AddDate(0, 0, 1)
	}

	buff := &bytes.Buffer{}
	csvWriter := json2csv.NewCSVWriter(buff)
	csvWriter.HeaderStyle = json2csv.DotNotationStyle

	csvWriter.WriteCSV(rows)

	tradesFileBytes := telegramBot.FileBytes{
		Name:  "data.csv",
		Bytes: buff.Bytes(),
	}
	docMsg := telegramUtils.GenerateReplyDocument(update, tradesFileBytes)
	bot.Send(docMsg)
}

func GetHistoricalData(bot *telegramBot.BotAPI, update telegramBot.Update) error {
	argString := update.Message.CommandArguments()
	argList := strings.Split(argString, "\n")
	fmt.Println(argList)
	if len(argList) != 4 {
		return fmt.Errorf("invalid arguments")
	}
	instrument := argList[0]
	fromDate := argList[1]
	toDate := argList[2]
	resolution := argList[3]

	fmt.Println(instrument, fromDate, toDate, resolution)

	isInstrumentValid, err := fyersSDK.IsValidInstrument(instrument)
	if err != nil {
		return err
	}

	if !isInstrumentValid {
		return fmt.Errorf("invalid instrument")
	}

	if _, err := time.Parse(time.RFC3339, fromDate); fromDate == "" || err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339, toDate); toDate == "" || err != nil {
		return err
	}
	if _, err := strconv.Atoi(resolution); err != nil {
		return fmt.Errorf("invalid resolutionString")
	}

	getHistoricalData(bot, update, instrument, fromDate, toDate, resolution)

	return nil
}
