package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

func handle(ctx context.Context) {
	lc, _ := lambdacontext.FromContext(ctx)
	slog.Info("Lambda Handler",
		"RequestID",
		lc.AwsRequestID,
		"FunctionArn",
		lc.InvokedFunctionArn)
}

func main() {
	lambda.Start(handle)
}
