package streams

import (
	"context"
	"dunno/internal/log"
	"errors"

	"github.com/aws/aws-lambda-go/events"
)

type BookIndexRecord struct {
	Id      string
	Title   string
	ISBN    string
	Authors []string
}

func stringList(av events.DynamoDBAttributeValue) []string {
	l := av.List()
	result := make([]string, len(l))
	for i, v := range l {
		result[i] = v.String()
	}
	return result
}

func Handle(_ context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		log.Logger.Infow("Processing DynamoDBEventRecord", "record", record)
		newImage := record.Change.NewImage
		if newImage == nil {
			return errors.New("NewImage field expected")
		}
		indexRecord := BookIndexRecord{
			Id:      newImage["id"].String(),
			Title:   newImage["title"].String(),
			ISBN:    newImage["isbn"].String(),
			Authors: stringList(newImage["authors"]),
		}
		log.Logger.Infow("BookIndexRecord created", "record", indexRecord)
	}
	return nil
}
