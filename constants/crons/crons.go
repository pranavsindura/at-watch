package cronConstants

const (
	Maintenance             string = "MAINTENANCE"
	StartMarket             string = "START_MARKET"
	StopMarket              string = "STOP_MARKET"
	UpdateOpenTradesInMongo string = "UPDATE_OPEN_TRADES_IN_MONGO"
)

const (
	CronMaintenance             string = "0 0 * * *"   // 0000 every day
	CronStartMarket             string = "0 9 * * *"   // 0900 every day
	CronStopMarket              string = "40 15 * * *" // 1540 every day
	CronUpdateOpenTradesInMongo string = "*/5 * * * *" // every 5th minute
)
