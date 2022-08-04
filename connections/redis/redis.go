package redisClient

import (
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
)

var redisClient *redis.Client

func Init() {
	DB, _ := strconv.Atoi(os.Getenv(envConstants.RedisDB))
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv(envConstants.RedisHost),
		Password: os.Getenv(envConstants.RedisPassword),
		DB:       DB,
	})
}

func Client() *redis.Client {
	return redisClient
}
