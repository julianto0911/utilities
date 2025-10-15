package utilities

import (
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"golang.org/x/crypto/bcrypt"
)

func HashBcrypt(password string, strength int) (string, error) {
	if strength == 0 {
		return password, nil
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), strength)
	return string(bytes), err
}

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

func GetAuthHeader(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	if token == "" {
		return ""
	}

	//retrieve token from header
	token = strings.TrimPrefix(token, "Bearer ")
	return token
}
