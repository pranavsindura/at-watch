package positionalStrategyModel

import (
	mongoClient "github.com/pranavsindura/at-watch/connections/mongo"
	mongoConstants "github.com/pranavsindura/at-watch/constants/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PositionalStrategy struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"userID"`
	InstrumentID primitive.ObjectID `bson:"instrumentID"`
	TimeFrame    string             `bson:"timeFrame"`
	IsActive     bool               `bson:"isActive"`
}

var PositionalStrategyCollection *mongo.Collection = nil

func GetPositionalStrategyCollection() *mongo.Collection {
	if PositionalStrategyCollection == nil {
		PositionalStrategyCollection = mongoClient.Client().Database(mongoConstants.DB).Collection(mongoConstants.PositionalStrategyCollection)
	}
	return PositionalStrategyCollection
}
