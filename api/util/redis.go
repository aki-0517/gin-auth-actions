package util

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func RedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	_, err := rdb.Ping(rdb.Context()).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to redis: %v", err))
	}
	return rdb
}
