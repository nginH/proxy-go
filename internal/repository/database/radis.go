package database

import (
	"context"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

func NewRedisClient() *RedisClient {

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	pass := os.Getenv("REDIS_PASSWORD")
	if pass == "" {
		pass = ""
	}
	db := os.Getenv("REDIS_DB")
	if db == "" {
		db = "0"
	}
	dbNum, err := strconv.Atoi(db)
	if err != nil {
		dbNum = 0
	}

	uri := host + ":" + port

	client := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: pass,
		DB:       dbNum,
	})

	return &RedisClient{
		Client: client,
		Ctx:    context.Background(),
	}
}
