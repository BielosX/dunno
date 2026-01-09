package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(_ context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		slog.Info("Processing SQS Event", "messageId", record.MessageId, "body", record.Body)
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
