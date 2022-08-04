package telegramHelpers

import (
	"context"

	"github.com/go-redis/redis/v8"
	cache "github.com/pranavsindura/at-watch/cache"
	instrumentModel "github.com/pranavsindura/at-watch/models/instrument"
	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	cacheTypes "github.com/pranavsindura/at-watch/types/cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DoesFyersAccessTokenExist() bool {
	return fyersSDK.GetFyersAccessToken() != ""
}

func GetUserSession(telegramUserID int64) (*cacheTypes.UserSession, error) {
	userSession, err := cache.UserSession(telegramUserID)
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if userSession != nil {
		return userSession, nil
	}
	coll := telegramUserModel.GetTelegramUserCollection()
	res := coll.FindOne(context.Background(), bson.M{
		"telegramUserID": telegramUserID,
	})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	user := &telegramUserModel.TelegramUserModel{}
	err = res.Decode(user)
	if err != nil {
		return nil, err
	}

	userSession = &cacheTypes.UserSession{
		AccessLevel: user.AccessLevel,
		UserID:      user.ID,
	}

	cache.SetUserSession(telegramUserID, *userSession)
	return userSession, nil
}

func GetChatIDByUserID(userID primitive.ObjectID) (int64, error) {
	chatID, err := cache.ChatIDByUserID(userID)
	if err != nil && err != redis.Nil {
		return -1, err
	}
	if chatID != -1 {
		return chatID, nil
	}
	coll := telegramUserModel.GetTelegramUserCollection()
	res := coll.FindOne(context.Background(), bson.M{
		"_id": userID,
	})
	err = res.Err()
	if err != nil {
		return -1, err
	}
	user := &telegramUserModel.TelegramUserModel{}
	err = res.Decode(user)
	if err != nil {
		return -1, err
	}

	chatID = user.TelegramChatID

	cache.SetChatIDByUserID(userID, chatID)
	return chatID, nil
}

func GetInstruments() ([]string, error) {
	instruments, err := cache.Instruments()
	if err != nil && err != redis.Nil {
		return make([]string, 0), err
	}
	if len(instruments) > 0 {
		return instruments, nil
	}

	coll := instrumentModel.GetInstrumentCollection()
	cur, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return make([]string, 0), err
	}

	instruments = make([]string, 0)

	for cur.Next(context.Background()) {
		var instrument instrumentModel.InstrumentModel
		err := cur.Decode(&instrument)
		if err != nil {
			return make([]string, 0), err
		}
		instruments = append(instruments, instrument.Symbol)
	}

	cache.SetInstruments(instruments)
	return instruments, nil
}
