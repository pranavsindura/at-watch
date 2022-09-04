package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pranavsindura/at-watch/cache"
	mongoClient "github.com/pranavsindura/at-watch/connections/mongo"
	redisClient "github.com/pranavsindura/at-watch/connections/redis"
	routerClient "github.com/pranavsindura/at-watch/connections/router"
	telegramClient "github.com/pranavsindura/at-watch/connections/telegram"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
	"github.com/pranavsindura/at-watch/crons"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
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

	fyersAccessToken, err := cache.FyersAccessToken()
	if err == nil {
		fyersSDK.SetFyersAccessToken(fyersAccessToken)
	}

	go herokuKeepAlive()

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
