package log

import (
	"dunno/api/config"

	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func init() {
	cfg := zap.NewProductionConfig()
	level, err := zap.ParseAtomicLevel(config.AppConfig.LogLevel)
	if err != nil {
		panic(err.Error())
	}
	cfg.Level = level
	logger, err := cfg.Build()
	if err != nil {
		panic(err.Error())
	}
	Logger = logger.Sugar()
}
