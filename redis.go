package utilities

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// this is useful for testing, to predefined behavior of the response

type Redis interface {
	Ping() error
	Get(name string) (string, error)
	Set(name string, value string) error
	SetWithDuration(name string, value string, d time.Duration) error
	Delete(name string) error
	GetKeysWithParam(name string) ([]string, error)
	PrintKeys()
}

func NewRedis(rdc *redis.Client, prefix string, expiracy int) Redis {
	return &rds{
		rdb:      rdc,
		expiracy: time.Duration(expiracy) * time.Second,
		prefix:   prefix,
	}
}

type rds struct {
	rdb      *redis.Client
	expiracy time.Duration
	prefix   string
}

func (c *rds) PrintKeys() {
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = c.rdb.Scan(context.Background(), cursor, "", 0).Result()
		if err != nil {
			panic(err)
		}

		for _, key := range keys {
			fmt.Println("key", key)
		}

		if cursor == 0 { // no more keys
			break
		}
	}
}

func (c *rds) SetWithDuration(name string, value string, d time.Duration) error {
	return c.rdb.Set(context.Background(), c.prefix+"_"+name, value, d).Err()
}

func (c *rds) Set(name string, value string) error {
	return c.rdb.Set(context.Background(), c.prefix+"_"+name, value, c.expiracy).Err()
}

func (c *rds) Get(name string) (string, error) {
	return c.rdb.Get(context.Background(), c.prefix+"_"+name).Result()
}

func (c *rds) Delete(name string) error {
	return c.rdb.Del(context.Background(), c.prefix+"_"+name).Err()
}

func (c *rds) Ping() error {
	return c.rdb.Ping(context.Background()).Err()
}

func (c *rds) GetKeysWithParam(name string) ([]string, error) {
	return c.rdb.Keys(context.Background(), c.prefix+"_"+name).Result()
}
