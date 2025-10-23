package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {

	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	var config zap.Config

	if atomicLevel.Level() == zapcore.DebugLevel {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = atomicLevel

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
