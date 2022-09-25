package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/pranavsindura/at-watch/cache"
	mongoClient "github.com/pranavsindura/at-watch/connections/mongo"
	redisClient "github.com/pranavsindura/at-watch/connections/redis"
	routerClient "github.com/pranavsindura/at-watch/connections/router"
	telegramClient "github.com/pranavsindura/at-watch/connections/telegram"
	"github.com/pranavsindura/at-watch/constants"
	cronConstants "github.com/pranavsindura/at-watch/constants/crons"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
	marketConstants "github.com/pranavsindura/at-watch/constants/market"
	"github.com/pranavsindura/at-watch/crons"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	"github.com/pranavsindura/at-watch/telegram"
	"github.com/rs/zerolog/log"
)

func main() {
	initLogger()
	initEnv()
	crons.Init()
	mongoClient.Init()
	telegramClient.Init()
	telegram.FetchUpdatesAndHandle()
	redisClient.Init()

	go herokuKeepAlive()

	go attemptAutoLoginAndStartMarket()

	// Blocks all logs, init at the end
	routerClient.Init()
}

func initLogger() {
	// TODO: Add Logs to File with Rotation
	log.Info().Msg("init logger")
	log.Logger = log.Logger.With().Caller().Logger()
	log.Info().Msg("logger file init done")
}

func initEnv() {
	log.Info().Msg("init .env")
	err := godotenv.Load()

	if err != nil {
		log.Fatal().Err(err)
	}
}

func herokuKeepAlive() {
	for {
		time.Sleep(time.Minute * 5)
		req, err := http.NewRequest(http.MethodGet, os.Getenv(envConstants.MyURL)+"/public/ping", nil)
		if err != nil {
			fmt.Println("unable to create request", err)
			continue
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("error in response", err)
			continue
		}
		fmt.Println("keep alive")
	}
}

func attemptAutoLoginAndStartMarket() {
	err := attemptAutoLogin()
	if err != nil {
		return
	}
	attemptStartMarket()
}

func attemptAutoLogin() error {
	fyersAccessToken, err := cache.FyersAccessToken()
	if err == nil {
		fyersSDK.SetFyersAccessToken(fyersAccessToken)
		notifications.Broadcast(constants.AccessLevelAdmin, "Successfully set Fyers Access token from Cache")
	} else {
		time.Sleep(time.Second * 5) // wait for router to init
		ok, err := fyersSDK.AutomateAdminLogin()
		if ok {
			fmt.Println("auto login successful")
			notifications.Broadcast(constants.AccessLevelAdmin, "Admin Auto Login successful")
		} else {
			fmt.Println("auto login unsuccessful")
			notifications.Broadcast(constants.AccessLevelAdmin, "Admin Auto Login failed\n\n"+err.Error())
			return err
		}
	}
	return nil
}

func attemptStartMarket() {
	now := time.Now()
	nowMinutes := now.Hour()*60 + now.Minute()
	marketOpen := marketConstants.MarketOpenHours*60 + marketConstants.MarketOpenMinutes
	marketClose := marketConstants.MarketCloseHours*60 + marketConstants.MarketCloseMinutes
	if nowMinutes >= marketOpen && nowMinutes < marketClose {
		_, err := marketSDK.Start()

		if err != nil {
			notifications.Broadcast(constants.AccessLevelAdmin, "Auto StartMarket failed\n\n"+err.Error())
			return
		}
		crons.MarketCron().Start()
		notifications.Broadcast(constants.AccessLevelUser, "Market has now Started")
	} else {
		notifications.Broadcast(constants.AccessLevelAdmin, "Auto StartMarket conditions not satisfied, time must be >="+strconv.Itoa(marketConstants.MarketOpenHours)+":"+strconv.Itoa(marketConstants.MarketOpenMinutes)+" and <"+strconv.Itoa(marketConstants.MarketCloseHours)+":"+strconv.Itoa(marketConstants.MarketCloseMinutes))
		// schedule a job to run start market
		id, _ := crons.ServerCron().AddFunc(cronConstants.CronStartMarket, crons.StartMarket)
		fmt.Println("added StartMarket cron to server", cronConstants.CronStartMarket, "ID", id)
		notifications.Broadcast(constants.AccessLevelAdmin, "Scheduled a cron for StartMarket")
	}
}
