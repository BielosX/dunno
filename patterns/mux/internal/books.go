package internal

import (
	"dunno/internal/global"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.uber.org/zap"
)

type BookResponse struct {
	Id      string   `json:"id"`
	Title   string   `json:"title"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
	Pages   int      `json:"pages"`
}

type BookRequest struct {
	Title   string   `json:"title"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
	Pages   int      `json:"pages"`
}

const (
	DynamoDbTag          = "dynamodb"
	DynamoDbPartitionKey = "partitionKey"
)

type BookRecord struct {
	Id      string `dynamodb:"partitionKey"`
	Title   string
	ISBN    string
	Authors []string
	Pages   int
}

func BooksKey(bookId string) map[string]types.AttributeValue {
	var record BookRecord
	primaryKeyName := GetFieldsByTagValue(&record, DynamoDbTag, DynamoDbPartitionKey)[0].Name
	return map[string]types.AttributeValue{
		primaryKeyName: &types.AttributeValueMemberS{
			Value: bookId,
		},
	}
}

func GetBook(r *Request[Unit]) *Response[BookResponse] {
	bookId := r.PathParams["id"]
	global.Logger.Info("Fetching book with ID", zap.String("id", bookId))
	out, err := global.DynamoDBClient.GetItem(r.Context, &dynamodb.GetItemInput{
		TableName: aws.String(global.AppConfig.BooksTableArn),
		Key:       BooksKey(bookId),
	})
	if err != nil {
		return ErrorResponse[BookResponse](err, http.StatusInternalServerError)
	}
	if out.Item == nil {
		return NotFoundResponse[BookResponse]()
	}
	var bookRecord BookRecord
	err = attributevalue.UnmarshalMap(out.Item, &bookRecord)
	if err != nil {
		return ErrorResponse[BookResponse](err, http.StatusInternalServerError)
	}
	return SuccessResponse(&BookResponse{
		Id:      bookRecord.Id,
		Title:   bookRecord.Title,
		ISBN:    bookRecord.ISBN,
		Authors: bookRecord.Authors,
		Pages:   bookRecord.Pages,
	})
}

func SaveBook(r *Request[BookRequest]) *Response[BookResponse] {
	id := uuid.New().String()
	record := BookRecord{
		Id:      id,
		Title:   r.Body.Title,
		ISBN:    r.Body.ISBN,
		Authors: r.Body.Authors,
		Pages:   r.Body.Pages,
	}
	av, err := attributevalue.MarshalMap(record)
	if err != nil {
		return ErrorResponse[BookResponse](err, http.StatusInternalServerError)
	}
	_, err = global.DynamoDBClient.PutItem(r.Context, &dynamodb.PutItemInput{
		TableName: aws.String(global.AppConfig.BooksTableArn),
		Item:      av,
	})
	if err != nil {
		return ErrorResponse[BookResponse](err, http.StatusInternalServerError)
	}
	return SuccessResponse(&BookResponse{
		Id:      record.Id,
		Title:   record.Title,
		ISBN:    record.ISBN,
		Authors: record.Authors,
		Pages:   record.Pages,
	})
}

func init() {
	RegisterFunc(global.Router, "/books/{id}", "GET", GetBook)
	RegisterFuncMatchContentType(global.Router, "/books", "POST", SaveBook, "application/json")
}
