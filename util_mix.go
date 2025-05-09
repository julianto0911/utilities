package utilities

import (
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
)

func ShortUUID() string {
	return shortuuid.New()
}

func NewUUID() string {
	return uuid.New().String()
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

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyz")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
