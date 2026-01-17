package books

import (
	"dunno/internal/api/config"
	clients "dunno/internal/aws"
	"dunno/internal/log"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

var allowedApplicationJson = middleware.AllowContentType("application/json")

func NewRouter() chi.Router {
	router := chi.NewRouter()
	router.Get(fmt.Sprintf("/{bookId}"), getBookById)
	router.With(allowedApplicationJson).Post("/", saveBook)
	return router
}

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

type GetBookResponse struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	ISBN        string    `json:"isbn"`
	Authors     []string  `json:"authors"`
	Pages       int       `json:"pages"`
	PublishDate time.Time `json:"publishDate"`
}

type SaveBookRequest struct {
	Title       string    `json:"title"`
	ISBN        string    `json:"isbn"`
	Authors     []string  `json:"authors"`
	Pages       int       `json:"pages"`
	PublishDate time.Time `json:"publishDate"`
}

func getBookById(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")
	key := map[string]string{
		"id": bookId,
	}
	av, err := attributevalue.MarshalMap(&key)
	if err != nil {
		log.Logger.Errorf("Unable to marshall key, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	result, err := clients.DynamoDbClient.GetItem(r.Context(), &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(config.ApiConfig.BooksTableArn),
		Key:            av,
	})
	if err != nil {
		log.Logger.Errorf("Unable to get dynamodb item, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.Item == nil || len(result.Item) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var record BookRecord
	err = attributevalue.UnmarshalMap(result.Item, &record)
	if err != nil {
		log.Logger.Errorf("Unable to unmarshall dynamodb record, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	publishDate, err := time.Parse(time.RFC3339, record.PublishDate)
	if err != nil {
		log.Logger.Errorf("Unable to parse publishDate, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := GetBookResponse{
		Id:          record.Id,
		Title:       record.Title,
		ISBN:        record.ISBN,
		Authors:     record.Authors,
		Pages:       record.Pages,
		PublishDate: publishDate,
	}
	out, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(out)
}

func saveBook(w http.ResponseWriter, r *http.Request) {
	log.Logger.Infof("Received Content-Type: %s", r.Header.Get("Content-Type"))
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Logger.Errorf("Unable to read request body, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var saveRequest SaveBookRequest
	err = json.Unmarshal(rawBody, &saveRequest)
	if err != nil {
		log.Logger.Errorf("Unable to unmarshall SaveBookRequest, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	now := time.Now().Format(time.RFC3339)
	bookId := uuid.New().String()
	publishDate := saveRequest.PublishDate
	record := BookRecord{
		Id:          bookId,
		Title:       saveRequest.Title,
		ISBN:        saveRequest.ISBN,
		Authors:     saveRequest.Authors,
		Pages:       saveRequest.Pages,
		PublishDate: publishDate.Format(time.RFC3339),
		Created:     now,
		Updated:     now,
	}
	av, err := attributevalue.MarshalMap(&record)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = clients.DynamoDbClient.PutItem(r.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(config.ApiConfig.BooksTableArn),
		Item:      av,
	})
	if err != nil {
		log.Logger.Errorf("Unable to PutItem into dynamodb table, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := &GetBookResponse{
		Id:          record.Id,
		Title:       record.Title,
		ISBN:        record.ISBN,
		Authors:     record.Authors,
		Pages:       record.Pages,
		PublishDate: publishDate,
	}
	out, err := json.Marshal(&response)
	if err != nil {
		log.Logger.Errorf("Unable to marshall GetBookResponse, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(out)
}
