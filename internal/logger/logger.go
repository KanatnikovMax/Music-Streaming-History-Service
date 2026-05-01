package logger

import (
	"MusicStreamingHistoryService/internal/config"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func MustBuild(cfg config.LoggerConfig) *zap.Logger {
	var logger *zap.Logger
	var err error

	switch cfg.Env {
	case "production":
		logger, err = zap.NewProduction()
	default:
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	return logger
}

func buildDevelopmentLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	return cfg.Build()
}

func buildProductionLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg.Build()
}
