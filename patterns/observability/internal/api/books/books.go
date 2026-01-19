package books

import (
	"dunno/internal/api/config"
	clients "dunno/internal/aws"
	"dunno/internal/log"
	"dunno/internal/opensearch"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

var allowedApplicationJson = middleware.AllowContentType("application/json")

const publishDateLayout = time.RFC3339

func NewRouter() chi.Router {
	router := chi.NewRouter()
	router.Get("/{bookId}", getBookById)
	router.Delete("/{bookId}", deleteBook)
	router.Get("/", listBooks)
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

type ListBooksResponse struct {
	Books   []GetBookResponse `json:"books"`
	LastKey string            `json:"lastKey,omitempty"`
}

type SaveBookRequest struct {
	Title       string    `json:"title"`
	ISBN        string    `json:"isbn"`
	Authors     []string  `json:"authors"`
	Pages       int       `json:"pages"`
	PublishDate time.Time `json:"publishDate"`
}

func bookKey(bookId string) map[string]types.AttributeValue {
	key := map[string]string{
		"id": bookId,
	}
	av, err := attributevalue.MarshalMap(&key)
	if err != nil {
		panic(err.Error())
	}
	return av
}

func getBookById(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")
	av := bookKey(bookId)
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
	publishDate, err := time.Parse(publishDateLayout, record.PublishDate)
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
	now := time.Now().Format(publishDateLayout)
	bookId := uuid.New().String()
	publishDate := saveRequest.PublishDate
	record := BookRecord{
		Id:          bookId,
		Title:       saveRequest.Title,
		ISBN:        saveRequest.ISBN,
		Authors:     saveRequest.Authors,
		Pages:       saveRequest.Pages,
		PublishDate: publishDate.Format(publishDateLayout),
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

const (
	IndexName = "books"
)

func deleteBook(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")
	av := bookKey(bookId)
	_, err := clients.DynamoDbClient.DeleteItem(r.Context(), &dynamodb.DeleteItemInput{
		TableName: aws.String(config.ApiConfig.BooksTableArn),
		Key:       av,
	})
	if err != nil {
		log.Logger.Errorf("Unable to delete DynamoDB item, error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	title := query.Get("title")
	isbn := query.Get("isbn")
	lastKey := query.Get("lastKey")
	authors := query["author"]
	size := int32(50)
	sizeParam := query.Get("size")
	if sizeParam != "" {
		p, err := strconv.ParseInt(sizeParam, 10, 32)
		if err != nil || p > 50 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		size = int32(p)
	}
	request := SearchRequest{
		Size: size,
		Sort: []map[string]any{
			{"id": "asc"},
		},
		SearchAfter: []string{},
		Query: QueryPayload{
			Bool: BoolPayload{
				Filter: []map[string]any{},
			},
		},
	}
	if lastKey != "" {
		request.SearchAfter = []string{lastKey}
	}
	if title != "" {
		request.Query.Bool.appendPrefixFilter(map[string]any{
			"title": strings.ToLower(title),
		})
	}
	if isbn != "" {
		request.Query.Bool.appendTermFilter(map[string]any{
			"isbn": isbn,
		})
	}
	if len(authors) > 0 {
		lowerAuthors := make([]string, len(authors))
		for i, v := range authors {
			lowerAuthors[i] = strings.ToLower(v)
		}
		request.Query.Bool.appendTermFilter(map[string]any{
			"authors": lowerAuthors,
		})
	}
	resp, err := opensearch.Client.Search(r.Context(), &opensearchapi.SearchReq{
		Indices: []string{IndexName},
		Body:    opensearchutil.NewJSONReader(&request),
	})
	if err != nil {
		log.Logger.Errorf("OpenSearch query failed, error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: add metric
	log.Logger.Infof("Search finished, took: %d ms", resp.Took)
	resultLastKey := ""
	hits := resp.Hits.Hits
	lenHits := len(hits)
	keys := make([]map[string]types.AttributeValue, lenHits)
	for i, v := range hits {
		if i == lenHits-1 && int32(lenHits) == size {
			resultLastKey = v.Sort[0].(string)
		}
		keys[i] = bookKey(v.ID)
	}
	lenKeys := len(keys)
	log.Logger.Infof("Found %d index entries", lenKeys)
	var bookRecords []BookRecord
	if lenKeys > 0 {
		out, err := clients.DynamoDbClient.BatchGetItem(r.Context(), &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				config.ApiConfig.BooksTableArn: {
					Keys: keys,
				},
			},
		})
		if err != nil {
			log.Logger.Errorf("DynamoDB BatchGetItem failed, error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tableEntries, ok := out.Responses[config.ApiConfig.BooksTableArn]
		if ok {
			_ = attributevalue.UnmarshalListOfMaps(tableEntries, &bookRecords)
		} else {
			log.Logger.Infof("Nothing found for table %s", config.ApiConfig.BooksTableArn)
		}
	}
	books := make([]GetBookResponse, len(bookRecords))
	for i, v := range bookRecords {
		publishDate, _ := time.Parse(publishDateLayout, v.PublishDate)
		books[i] = GetBookResponse{
			Id:          v.Id,
			Title:       v.Title,
			ISBN:        v.ISBN,
			Authors:     v.Authors,
			Pages:       v.Pages,
			PublishDate: publishDate,
		}
	}
	booksResponse := ListBooksResponse{
		Books:   books,
		LastKey: resultLastKey,
	}
	response, err := json.Marshal(&booksResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(response)
}
