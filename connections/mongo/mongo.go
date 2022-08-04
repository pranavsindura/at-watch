package mongoClient

import (
	"context"
	"os"
	"time"

	envConstants "github.com/pranavsindura/at-watch/constants/env"
	mongoConstants "github.com/pranavsindura/at-watch/constants/mongo"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func Init() {
	log.Info().Msg("init mongo")
	uri := os.Getenv(envConstants.MongoURI)
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal().Err(err)
	}

	mongoClient = client
	addIndexes(mongoClient)
}

func Client() *mongo.Client {
	return mongoClient
}

func addIndexes(mongoClient *mongo.Client) {
	// TelegramUserIndex
	telegramUserIndex := mongo.IndexModel{
		Keys: bson.M{
			"telegramUserID": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	mongoClient.Database(mongoConstants.DB).Collection(mongoConstants.TelegramUserCollection).Indexes().CreateOne(context.Background(), telegramUserIndex)

	// Instrument Index
	instrumentIndex := mongo.IndexModel{
		Keys: bson.M{
			"symbol": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	mongoClient.Database(mongoConstants.DB).Collection(mongoConstants.InstrumentCollection).Indexes().CreateOne(context.Background(), instrumentIndex)
}
