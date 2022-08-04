package telegramUserModel

import (
	mongoClient "github.com/pranavsindura/at-watch/connections/mongo"
	mongoConstants "github.com/pranavsindura/at-watch/constants/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TelegramUserModel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	FirstName      string             `bson:"firstName"`
	LastName       string             `bson:"lastName"`
	TelegramUserID int64              `bson:"telegramUserID"`
	TelegramChatID int64              `bson:"telegramChatID"`
	AccessLevel    int                `bson:"accessLevel"`
}

var TelegramUserCollection *mongo.Collection = nil

func GetTelegramUserCollection() *mongo.Collection {
	if TelegramUserCollection == nil {
		TelegramUserCollection = mongoClient.Client().Database(mongoConstants.DB).Collection(mongoConstants.TelegramUserCollection)
	}
	return TelegramUserCollection
}
