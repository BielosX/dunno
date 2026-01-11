package internal

import (
	"dunno/internal/global"

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

type ListBooksResponse struct {
	Books            []BookResponse `json:"books"`
	LastEvaluatedKey *string        `json:"lastEvaluatedKey,omitempty"`
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
		return ServerError[BookResponse](err)
	}
	if out.Item == nil {
		return NotFoundResponse[BookResponse]()
	}
	var bookRecord BookRecord
	err = attributevalue.UnmarshalMap(out.Item, &bookRecord)
	if err != nil {
		return ServerError[BookResponse](err)
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
		return ServerError[BookResponse](err)
	}
	_, err = global.DynamoDBClient.PutItem(r.Context, &dynamodb.PutItemInput{
		TableName: aws.String(global.AppConfig.BooksTableArn),
		Item:      av,
	})
	if err != nil {
		return ServerError[BookResponse](err)
	}
	return SuccessResponse(&BookResponse{
		Id:      record.Id,
		Title:   record.Title,
		ISBN:    record.ISBN,
		Authors: record.Authors,
		Pages:   record.Pages,
	})
}

const (
	lastEvaluatedKeyQueryParam = "LastEvaluatedKey"
	limitQueryParam            = "limit"
)

func ListBooks(r *Request[Unit]) *Response[ListBooksResponse] {
	encodedLastEvaluatedKey := GetQueryParam(r, lastEvaluatedKeyQueryParam)
	var limit *int32
	var err error
	limitStr := GetQueryParam(r, limitQueryParam)
	if limitStr != "" {
		limit, err = ParseInt32(limitStr)
		if err != nil {
			return ServerError[ListBooksResponse](err)
		}
	}
	var lastEvaluatedKey map[string]types.AttributeValue
	if encodedLastEvaluatedKey != "" {
		lastEvaluatedKey, err = DecodeLastEvaluatedKey(encodedLastEvaluatedKey)
		if err != nil {
			return ServerError[ListBooksResponse](err)
		}
	}
	out, err := global.DynamoDBClient.Scan(r.Context, &dynamodb.ScanInput{
		Limit:             limit,
		TableName:         aws.String(global.AppConfig.BooksTableArn),
		ExclusiveStartKey: lastEvaluatedKey,
	})
	if err != nil {
		return ServerError[ListBooksResponse](err)
	}
	lastEvaluatedKey = out.LastEvaluatedKey
	var lastEvaluatedKeyPtr *string
	if out.LastEvaluatedKey != nil {
		result, err := EncodeLastEvaluatedKey(lastEvaluatedKey)
		if err != nil {
			return ServerError[ListBooksResponse](err)
		}
		lastEvaluatedKeyPtr = &result
	}
	var books []BookRecord
	err = attributevalue.UnmarshalListOfMaps(out.Items, &books)
	if err != nil {
		return ServerError[ListBooksResponse](err)
	}
	items := make([]BookResponse, 0)
	for _, book := range books {
		items = append(items, BookResponse{
			Id:      book.Id,
			Title:   book.Title,
			ISBN:    book.ISBN,
			Authors: book.Authors,
			Pages:   book.Pages,
		})
	}
	return SuccessResponse(&ListBooksResponse{
		Books:            items,
		LastEvaluatedKey: lastEvaluatedKeyPtr,
	})
}

func init() {
	RegisterFunc(global.Router, "/books/{id}", "GET", GetBook)
	RegisterFunc(global.Router, "/books", "GET", ListBooks)
	RegisterFunc(global.Router, "/books", "POST", SaveBook)
	RegisterFuncMatchContentType(global.Router, "/books", "POST", SaveBook, "application/json")
}
