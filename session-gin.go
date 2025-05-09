package utilities

import (
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type GinSession struct {
}

func (gs *GinSession) Remove(c *gin.Context, key string) error {
	session := sessions.Default(c)
	session.Delete(key)
	return session.Save()
}

func (gs *GinSession) Clear(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}

func (gs *GinSession) GetString(c *gin.Context, key string) (string, error) {
	session := sessions.Default(c)
	value := session.Get(key)
	if value == nil {
		return "", errors.New("value not found")
	}

	return value.(string), nil
}

func (gs *GinSession) GetInt(c *gin.Context, key string) (int, error) {
	session := sessions.Default(c)
	value := session.Get(key)
	if value == nil {
		return 0, errors.New("value not found")
	}

	return value.(int), nil
}

func (gs *GinSession) GetInt64(c *gin.Context, key string) (int64, error) {
	session := sessions.Default(c)
	value := session.Get(key)
	if value == nil {
		return 0, errors.New("value not found")
	}

	return value.(int64), nil
}

func (gs *GinSession) GetFloat64(c *gin.Context, key string) (float64, error) {
	session := sessions.Default(c)
	value := session.Get(key)
	if value == nil {
		return 0, errors.New("value not found")
	}

	return value.(float64), nil
}

func (gs *GinSession) GetBool(c *gin.Context, key string) (bool, error) {
	session := sessions.Default(c)
	value := session.Get(key)
	if value == nil {
		return false, errors.New("value not found")
	}

	return value.(bool), nil
}

func (gs *GinSession) SetString(c *gin.Context, key string, value string) error {
	session := sessions.Default(c)
	session.Set(key, value)
	return session.Save()
}

func (gs *GinSession) SetInt(c *gin.Context, key string, value int) error {
	session := sessions.Default(c)
	session.Set(key, value)
	return session.Save()
}

func (gs *GinSession) SetInt64(c *gin.Context, key string, value int64) error {
	session := sessions.Default(c)
	session.Set(key, value)
	return session.Save()
}

func (gs *GinSession) SetFloat64(c *gin.Context, key string, value float64) error {
	session := sessions.Default(c)
	session.Set(key, value)
	return session.Save()
}

func (gs *GinSession) SetBool(c *gin.Context, key string, value bool) error {
	session := sessions.Default(c)
	session.Set(key, value)
	return session.Save()
}

func (gs *GinSession) Save(c *gin.Context) error {
	session := sessions.Default(c)
	return session.Save()
}
