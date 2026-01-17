package streams

import (
	"context"
	"dunno/internal/log"
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

type BookRecord struct {
	Id          string   `dynamodbav:"id"`
	Title       string   `dynamodbav:"title"`
	ISBN        string   `dynamodbav:"isbn"`
	Authors     []string `dynamodbav:"authors"`
	Pages       int      `dynamodbav:"pages"`
	PublishDate string   `dynamodbav:"publishDate"`
	Created     string   `dynamodbav:"created"`
	Updated     string   `dynamodbav:"updated"`
}

func Handle(_ context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		log.Logger.Infow("Processing DynamoDBEventRecord", "record", record)
		if record.Change.NewImage == nil {
			return errors.New("NewImage field expected")
		}
		b, err := json.Marshal(record.Change.NewImage)
		if err != nil {
			return err
		}
		av, err := attributevalue.UnmarshalMapJSON(b)
		if err != nil {
			return err
		}
		var bookRecord BookRecord
		err = attributevalue.UnmarshalMap(av, &bookRecord)
		if err != nil {
			return err
		}
		log.Logger.Infow("Processing BookRecord", "book", bookRecord)
	}
	return nil
}
