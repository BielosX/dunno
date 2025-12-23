package main

import (
	"github.com/caarlos0/env"
)

type Environ struct {
	MoviesTableArn string `env:"MOVIES_TABLE_ARN"`
}

var environ *Environ

func loadEnv() (*Environ, error) {
	var environ Environ
	err := env.Parse(&environ)
	if err != nil {
		return nil, err
	}
	return &environ, nil
}
