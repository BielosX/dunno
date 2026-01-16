package main

import (
	"dunno/api"
	"dunno/api/config"
	"dunno/api/log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v11"
)

const (
	apiGatewayHandler = "apiGatewayHandler"
)

func main() {
	var handlerConfig config.HandlerConfig
	err := env.Parse(&handlerConfig)
	if err != nil {
		log.Logger.Errorf("Unable to parse HandlerConfig, %v", err)
		os.Exit(1)
	}
	handler := handlerConfig.Handler
	switch handler {
	case apiGatewayHandler:
		lambda.Start(api.LambdaHandler)
	default:
		log.Logger.Errorf("Unable to find handler: %s", handler)
		os.Exit(1)
	}
}
