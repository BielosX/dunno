package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func init() {
	var err error
	environ, err = loadEnv()
	if err != nil {
		panic(err)
	}
	handlers = newHandlers()
	handlers.register("GetMovieById", GetMovieById)
	handlers.register("SaveMovie", SaveMovie)
	handlers.register("ListMovies", ListMovies)
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	dynamodbClient = dynamodb.NewFromConfig(cfg)
}

func main() {
	handlers.run()
}
