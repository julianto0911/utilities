package utilities

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"gopkg.in/natefinch/lumberjack.v2"
)

func MockLogs() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func NewLogger(path string, debug bool) (*zap.Logger, error) {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path + time.Now().Format("20060102") + ".log",
		MaxSize:    100, // megabytes
		MaxBackups: 30,
		MaxAge:     30, // days
	})

	pe := zap.NewProductionEncoderConfig()
	if debug {
		pe = zap.NewDevelopmentEncoderConfig()
	}

	fileEncoder := zapcore.NewJSONEncoder(pe)
	pe.TimeKey = "timestamp"
	pe.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(w), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	l := zap.New(core)

	return l, nil
}
