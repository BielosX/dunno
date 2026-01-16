package config

import (
	"github.com/caarlos0/env/v11"
)

type GlobalConfig struct {
	BooksTableArn string `env:"BOOKS_TABLE_ARN"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
}

type HandlerConfig struct {
	Handler string `env:"_HANDLER"`
}

var AppConfig GlobalConfig

func init() {
	err := env.Parse(&AppConfig)
	if err != nil {
		panic(err.Error())
	}
}
