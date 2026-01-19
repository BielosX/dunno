package streams

import (
	"context"
	"dunno/internal/log"
	"dunno/internal/opensearch"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

type BookIndexRecord struct {
	Id      string   `json:"id"`
	Title   string   `json:"title"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
}

const (
	BooksIndexName = "books"
)

func Handle(ctx context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		log.Logger.Infow("Processing DynamoDBEventRecord", "record", record)
		if record.EventName == string(events.DynamoDBOperationTypeRemove) {
			bookId := record.Change.Keys["id"].String()
			resp, err := opensearch.Client.Document.Delete(ctx, opensearchapi.DocumentDeleteReq{
				Index:      BooksIndexName,
				DocumentID: bookId,
			})
			if err != nil {
				return err
			}
			log.Logger.Infow("Deleted index", "entry", resp)
		} else {
			newImage := record.Change.NewImage
			if newImage == nil {
				return errors.New("NewImage field expected")
			}
			bookId := newImage["id"].String()
			authors := newImage["authors"].List()
			lowerAuthors := make([]string, len(authors))
			for i, a := range authors {
				lowerAuthors[i] = strings.ToLower(a.String())
			}
			indexRecord := BookIndexRecord{
				Id:      bookId,
				Title:   strings.ToLower(newImage["title"].String()),
				ISBN:    newImage["isbn"].String(),
				Authors: lowerAuthors,
			}
			log.Logger.Infow("BookIndexRecord created", "id", bookId, "record", indexRecord)
			resp, err := opensearch.Client.Index(ctx, opensearchapi.IndexReq{
				Index:      BooksIndexName,
				DocumentID: bookId,
				Body:       opensearchutil.NewJSONReader(&indexRecord),
				Params: opensearchapi.IndexParams{
					Refresh: "true",
				},
			})
			if err != nil {
				return err
			}
			log.Logger.Infow("Created index", "entry", resp)
		}
	}
	return nil
}
