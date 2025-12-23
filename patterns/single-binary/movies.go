package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type SaveMovieRequest struct {
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"releaseDate"`
}

type MovieRecord struct {
	Id          string    `dynamodbav:"id,string,pk"`
	Title       string    `dynamodbav:"title"`
	ReleaseDate time.Time `dynamodbav:"releaseDate"`
}

type GetMovieResponse struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"releaseDate"`
}

type ListMoviesResponse struct {
	Movies           []GetMovieResponse `json:"movies"`
	LastEvaluatedKey *string            `json:"LastEvaluatedKey,omitempty"`
}

func primaryKey(id string) map[string]types.AttributeValue {
	out, _ := attributevalue.MarshalMap(map[string]string{
		"id": id,
	})
	return out
}

func GetMovieById(ctx context.Context, request Request) (Response, error) {
	movieId, ok := request.PathParameters["movieId"]
	if !ok {
		return Response{
			StatusCode: http.StatusBadRequest,
			Body:       "movieId path param expected",
		}, nil
	}
	slog.Info("Fetching movie", "movieId", movieId)
	out, err := dynamodbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(environ.MoviesTableArn),
		Key:       primaryKey(movieId),
	})
	if err != nil {
		return InternalServerError(err), nil
	}
	if out.Item == nil || len(out.Item) == 0 {
		return Response{
			StatusCode: http.StatusNotFound,
		}, nil
	}
	var record MovieRecord
	err = attributevalue.UnmarshalMap(out.Item, &record)
	if err != nil {
		return InternalServerError(err), nil
	}
	response := &GetMovieResponse{
		Id:          record.Id,
		Title:       record.Title,
		ReleaseDate: record.ReleaseDate,
	}
	responseBody, err := json.Marshal(response)
	if err != nil {
		return InternalServerError(err), nil
	}
	return Response{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func SaveMovie(ctx context.Context, request Request) (Response, error) {
	slog.Info("Received", "headers", request.Headers)
	contentType := getHeader(request.Headers, "Content-Type")
	if contentType == nil || *contentType != "application/json" {
		return Response{
			StatusCode: http.StatusBadRequest,
			Body:       "application/json Content-Type expected",
		}, nil
	}
	body, err := unmarshallBody[SaveMovieRequest](&request)
	if err != nil {
		return Response{StatusCode: http.StatusBadRequest}, nil
	}
	record := &MovieRecord{
		Id:          uuid.New().String(),
		Title:       body.Title,
		ReleaseDate: body.ReleaseDate,
	}
	out, err := attributevalue.MarshalMap(record)
	if err != nil {
		return InternalServerError(err), nil
	}
	_, err = dynamodbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(environ.MoviesTableArn),
		Item:      out,
	})
	if err != nil {
		return InternalServerError(err), nil
	}
	responseBody, err := json.Marshal(record)
	if err != nil {
		return InternalServerError(err), nil
	}
	return Response{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}

func ListMovies(ctx context.Context, request Request) (Response, error) {
	exclusiveStartKeyParam := getOptional(request.QueryStringParameters, "exclusiveStartKey")
	limitParam := getOptional(request.QueryStringParameters, "limit")
	var limit *int32
	if limitParam != nil {
		val, err := parseInt32(*limitParam)
		if err != nil {
			return Response{StatusCode: http.StatusBadRequest}, nil
		}
		limit = &val
	}
	var exclusiveStartKey map[string]types.AttributeValue
	if exclusiveStartKeyParam != nil {
		var err error
		exclusiveStartKey, err = attributevalue.UnmarshalMapJSON([]byte(*exclusiveStartKeyParam))
		if err != nil {
			return InternalServerError(err), nil
		}
	}
	out, err := dynamodbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName:         aws.String(environ.MoviesTableArn),
		Limit:             limit,
		ExclusiveStartKey: exclusiveStartKey,
	})
	if err != nil {
		return InternalServerError(err), nil
	}
	lastEvaluatedKey, err := convertLastEvaluatedKey(out.LastEvaluatedKey)
	if err != nil {
		return InternalServerError(err), nil
	}
	var result ListMoviesResponse
	result.LastEvaluatedKey = lastEvaluatedKey
	result.Movies = make([]GetMovieResponse, len(out.Items))
	for i, item := range out.Items {
		var record MovieRecord
		err = attributevalue.UnmarshalMap(item, &record)
		if err != nil {
			return InternalServerError(err), nil
		}
		result.Movies[i] = GetMovieResponse{
			Id:          record.Id,
			Title:       record.Title,
			ReleaseDate: record.ReleaseDate,
		}
	}
	responseBody, err := json.Marshal(&result)
	if err != nil {
		return InternalServerError(err), nil
	}
	return Response{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}
