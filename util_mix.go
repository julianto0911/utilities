package utilities

import "github.com/redis/go-redis/v9"

func NewRedisClient(url, port, password string, dbIndex int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     url + ":" + port,
		Password: password,
		DB:       dbIndex,
	})
}
