package positionalTradeModel

import (
	mongoClient "github.com/pranavsindura/at-watch/connections/mongo"
	mongoConstants "github.com/pranavsindura/at-watch/constants/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PositionalSuperTrendData struct {
	Index               int       `bson:"index"`
	ATR                 float64   `bson:"atr"`
	PastTRList          []float64 `bson:"pastTRList"`
	BasicUpperBound     float64   `bson:"basicUpperBound"`
	BasicLowerBound     float64   `bson:"basicLowerBound"`
	FinalUpperBound     float64   `bson:"finalUpperBound"`
	FinalLowerBound     float64   `bson:"finalLowerBound"`
	SuperTrend          float64   `bson:"superTrend"`
	SuperTrendDirection bool      `bson:"superTrendDirection"`
	IsUsable            bool      `bson:"isUsable"`
}

type PositionalIndicators struct {
	SuperTrend PositionalSuperTrendData `bson:"superTrend"`
}

type PositionalCandle struct {
	TS         int64   `bson:"ts"`
	DateString string  `bson:"dateString"`
	Day        string  `bson:"day"`
	Open       float64 `bson:"open"`
	High       float64 `bson:"high"`
	Low        float64 `bson:"low"`
	Close      float64 `bson:"close"`
	Volume     float64 `bson:"volume"`
}

type PositionalCombinedCandle struct {
	Candle     PositionalCandle     `bson:"candle"`
	Indicators PositionalIndicators `bson:"indicators"`
}

type PositionalTrade struct {
	ID             primitive.ObjectID       `bson:"_id"`
	StrategyID     primitive.ObjectID       `bson:"strategyID"`
	Status         int                      `bson:"status"`
	StatusText     string                   `bson:"statusText"`
	TradeType      int                      `bson:"tradeType"`
	TradeTypeText  string                   `bson:"tradeTypeText"`
	Lots           int                      `bson:"lots"`
	Entry          PositionalCombinedCandle `bson:"entry"`
	PL             float64                  `bson:"PL"`
	Exit           PositionalCombinedCandle `bson:"exit"`
	ExitReason     int                      `bson:"exitReason"`
	ExitReasonText string                   `bson:"exitReasonText"`
	Brokerage      float64                  `bson:"brokerage"`
	UpdatedAt      int64                    `bson:"updatedAt"` // its possible that this field may not exist for all trades in db, newly added 4/9/22
}

var PositionalTradeCollection *mongo.Collection = nil

func GetPositionalTradeCollection() *mongo.Collection {
	if PositionalTradeCollection == nil {
		PositionalTradeCollection = mongoClient.Client().Database(mongoConstants.DB).Collection(mongoConstants.PositionalTradeCollection)
	}
	return PositionalTradeCollection
}
