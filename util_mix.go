package utilities

import (
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(url, port, password string, dbIndex int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     url + ":" + port,
		Password: password,
		DB:       dbIndex,
	})
}

func EnvArray(envName string) []string {
	val := os.Getenv(envName)
	ar := strings.Split(val, ",")
	return ar
}
func EnvInt(envName string) int {
	val, _ := strconv.Atoi(os.Getenv(envName))
	return val
}
func EnvBool(envName string) bool {
	val := os.Getenv(envName)
	return strings.ToUpper(val) == "TRUE"
}
func EnvString(envName string) string {
	return os.Getenv(envName)
}
