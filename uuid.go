package utilities

import (
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
)

func ShortUUID() string {
	return shortuuid.New()
}

func NewUUID() string {
	return uuid.New().String()
}
