package global

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	env "github.com/caarlos0/env/v11"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Config struct {
	BooksTableArn string `env:"BOOKS_TABLE_ARN"`
}

var Logger *zap.SugaredLogger
var Router *mux.Router
var DynamoDBClient *dynamodb.Client
var AppConfig *Config

func init() {
	prodLogger, _ := zap.NewProduction()
	Logger = prodLogger.Sugar()
	Router = mux.NewRouter()
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	DynamoDBClient = dynamodb.NewFromConfig(cfg)
	var appConfig Config
	err = env.Parse(&appConfig)
	AppConfig = &appConfig
	if err != nil {
		panic(err)
	}
}
