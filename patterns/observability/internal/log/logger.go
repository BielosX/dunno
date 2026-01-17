package log

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

var LoggerConfig Config

func init() {
	err := env.Parse(&LoggerConfig)
	if err != nil {
		panic(err.Error())
	}
	cfg := zap.NewProductionConfig()
	level, err := zap.ParseAtomicLevel(LoggerConfig.LogLevel)
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
