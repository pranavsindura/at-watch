package instrumentModel

import (
	mongoClient "github.com/pranavsindura/at-watch/connections/mongo"
	mongoConstants "github.com/pranavsindura/at-watch/constants/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type InstrumentModel struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Symbol string             `bson:"symbol"`
}

var InstrumentCollection *mongo.Collection = nil

func GetInstrumentCollection() *mongo.Collection {
	if InstrumentCollection == nil {
		InstrumentCollection = mongoClient.Client().Database(mongoConstants.DB).Collection(mongoConstants.InstrumentCollection)
	}
	return InstrumentCollection
}
