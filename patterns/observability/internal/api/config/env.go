package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	BooksTableArn string `env:"BOOKS_TABLE_ARN"`
}

var ApiConfig Config

func Load() {
	err := env.Parse(&ApiConfig)
	if err != nil {
		panic(err.Error())
	}
}
