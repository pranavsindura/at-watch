package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	redisClient "github.com/pranavsindura/at-watch/connections/redis"
	cacheTypes "github.com/pranavsindura/at-watch/types/cache"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ctx = context.Background()

func createKey(args []string) string {
	key := ""
	for idx, arg := range args {
		if idx > 0 {
			key += ":"
		}
		key += arg
	}
	return key
}

func ClearAll() (string, error) {
	clearResult := redisClient.Client().FlushAll(ctx)
	return clearResult.Result()
}

const userSessionPrefix = "USER_SESSION"

func createUserSessionKey(telegramUserID int64) string {
	return createKey([]string{userSessionPrefix, strconv.Itoa(int(telegramUserID))})
}
func UserSession(telegramUserID int64) (*cacheTypes.UserSession, error) {
	key := createUserSessionKey(telegramUserID)
	getResult := redisClient.Client().Get(ctx, key)
	result, err := getResult.Result()
	if err != nil {
		return nil, err
	}
	userSession := &cacheTypes.UserSession{}
	err = json.Unmarshal([]byte(result), userSession)
	if err != nil {
		return nil, err
	}
	return userSession, nil
}
func SetUserSession(telegramUserID int64, data cacheTypes.UserSession) (string, error) {
	key := createUserSessionKey(telegramUserID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	setResult := redisClient.Client().Set(ctx, key, jsonData, 0)
	return setResult.Result()
}
func DeleteUserSession(telegramUserID int64) (int64, error) {
	key := createUserSessionKey(telegramUserID)
	delResult := redisClient.Client().Del(ctx, key)
	return delResult.Result()
}

const chatIDByUserIDPrefix = "CHAT_ID"

func createChatIDByUserIDKey(userID primitive.ObjectID) string {
	return createKey([]string{chatIDByUserIDPrefix, userID.Hex()})
}
func ChatIDByUserID(userID primitive.ObjectID) (int64, error) {
	key := createChatIDByUserIDKey(userID)
	getResult := redisClient.Client().Get(ctx, key)
	result, err := getResult.Result()
	if err != nil {
		return -1, err
	}
	chatID, err := strconv.Atoi(result)
	if err != nil {
		return -1, err
	}
	return int64(chatID), nil
}
func SetChatIDByUserID(userID primitive.ObjectID, chatID int64) (string, error) {
	key := createChatIDByUserIDKey(userID)
	setResult := redisClient.Client().Set(ctx, key, chatID, 0)
	return setResult.Result()
}
func DeleteChatIDByUserID(userID primitive.ObjectID) (int64, error) {
	key := createChatIDByUserIDKey(userID)
	delResult := redisClient.Client().Del(ctx, key)
	return delResult.Result()
}

const fyersAccessTokenPrefix = "FYERS_ACCESS_TOKEN"

func createFyersAccessTokenKey(telegramUserID int64) string {
	return createKey([]string{fyersAccessTokenPrefix, strconv.Itoa(int(telegramUserID))})
}
func FyersAccessToken(telegramUserID int64) (string, error) {
	key := createFyersAccessTokenKey(telegramUserID)
	getResult := redisClient.Client().Get(ctx, key)
	return getResult.Result()
}
func SetFyersAccessToken(telegramUserID int64, token string) (string, error) {
	key := createFyersAccessTokenKey(telegramUserID)
	now := time.Now()
	nextDay6AM := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location())
	if nextDay6AM.Unix() < now.Unix() {
		// next day
		nextDay6AM = nextDay6AM.Add(time.Hour * 24)
	}
	expiration := nextDay6AM.Sub(now)
	setResult := redisClient.Client().Set(ctx, key, token, expiration)
	return setResult.Result()
}
func DeleteFyersAccessToken(telegramUserID int64) (int64, error) {
	key := createFyersAccessTokenKey(telegramUserID)
	delResult := redisClient.Client().Del(ctx, key)
	return delResult.Result()
}

const instrumentsPrefix = "INSTRUMENTS"

func createInstrumentsKey() string {
	return createKey([]string{instrumentsPrefix})
}
func Instruments() ([]string, error) {
	key := createInstrumentsKey()
	getResult := redisClient.Client().Get(ctx, key)
	results, err := getResult.Result()
	if err != nil {
		return make([]string, 0), err
	}
	instruments := make([]string, 0)
	json.Unmarshal([]byte(results), &instruments)
	return instruments, nil
}
func SetInstruments(instruments []string) (string, error) {
	key := createInstrumentsKey()
	instrumentsString, _ := json.Marshal(instruments)
	setResult := redisClient.Client().Set(ctx, key, instrumentsString, 0)
	return setResult.Result()
}
func DeleteInstruments() (int64, error) {
	key := createInstrumentsKey()
	delResult := redisClient.Client().Del(ctx, key)
	return delResult.Result()
}
