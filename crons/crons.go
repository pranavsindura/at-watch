package crons

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pranavsindura/at-watch/cache"
	"github.com/pranavsindura/at-watch/constants"
	cronConstants "github.com/pranavsindura/at-watch/constants/crons"
	positionalTradeModel "github.com/pranavsindura/at-watch/models/positionalTrade"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	marketSDK "github.com/pranavsindura/at-watch/sdk/market"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var serverCron *cron.Cron = nil
var marketCron *cron.Cron = nil

func ServerCron() *cron.Cron {
	if serverCron == nil {
		serverCron = cron.New()
	}
	return serverCron
}

func MarketCron() *cron.Cron {
	if marketCron == nil {
		marketCron = cron.New()
	}
	return marketCron
}

func Init() {
	id, _ := ServerCron().AddFunc(cronConstants.CronMaintenance, Maintenance)
	fmt.Println("added Maintenance cron to server", cronConstants.CronMaintenance, "ID", id)

	ServerCron().Start()

	id, _ = MarketCron().AddFunc(cronConstants.CronStopMarket, StopMarket)
	fmt.Println("added StopMarket cron to market", cronConstants.CronStopMarket, "ID", id)
	id, _ = MarketCron().AddFunc(cronConstants.CronUpdateOpenTradesInMongo, UpdateOpenTradesInMongo)
	fmt.Println("added UpdateOpenTradesInMongo cron to market", cronConstants.CronUpdateOpenTradesInMongo, "ID", id)
}

func Maintenance() {
	// maintenance logic also exists in /maintenance command
	notifications.Broadcast(constants.AccessLevelUser, "Server is performing nightly maintenance, please avoid any actions till maintenance is over")

	fyersSDK.SetFyersAccessToken("")
	cache.ClearAll()
	marketSDK.Stop()

	notifications.Broadcast(constants.AccessLevelUser, "Server has finished nightly maintenance")
}

func StopMarket() {
	MarketCron().Stop()
	_, err := marketSDK.Stop()

	if err != nil {
		fmt.Println("error while stopping market", err)
		notifications.Broadcast(constants.AccessLevelCreator, "error while stopping market\n\n"+err.Error())
		return
	}

	notifications.Broadcast(constants.AccessLevelUser, "Market has now Stopped")
}

func UpdateOpenTradesInMongo() {
	if !marketSDK.IsMarketWatchActive() {
		fmt.Println("market watch is not active right now, cannot UpdateOpenTradesInMongo")
		return
	}
	// Positional
	posTrades, _, err := marketSDK.GetAllPositionalOpenTrades()
	if err != nil {
		fmt.Println("error while updating positional trades in mongo", err)
		notifications.Broadcast(constants.AccessLevelCreator, "error while updating positional trades in mongo\n\n"+err.Error())
	} else {
		writeModels := []mongo.WriteModel{}
		for _, trade := range posTrades {
			writeModel := mongo.
				NewUpdateOneModel().
				SetFilter(bson.M{"_id": trade.ID}).
				SetUpdate(bson.M{
					"$set": bson.M{
						"PL":        trade.PL,
						"updatedAt": trade.UpdatedAtTS,
					},
				})
			writeModels = append(writeModels, writeModel)
		}
		res, err := positionalTradeModel.GetPositionalTradeCollection().BulkWrite(context.Background(), writeModels)
		if err != nil {
			fmt.Println("error while updating positional trades in mongo", err)
			notifications.Broadcast(constants.AccessLevelCreator, "error while updating positional trades in mongo - bulk write\n\n"+err.Error())
		} else {
			fmt.Println("updated "+strconv.Itoa(int(res.ModifiedCount))+" positional trades in mongo", err)
			// notifications.Broadcast(constants.AccessLevelCreator, "updated "+strconv.Itoa(int(res.ModifiedCount))+" positional trades in mongo")
		}
	}
}
