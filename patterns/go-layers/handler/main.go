package main

import (
	"context"
	"dunno/api"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"plugin"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var s3Client api.S3Client

const (
	S3ClientFactoryName = "NewS3Client"
)

func errorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{}, err
}

func uploadFile(ctx context.Context,
	bucket, key string,
	request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Uploading file to S3", "bucket", bucket, "key", key)
	contentType, ok := request.Headers["Content-Type"]
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusBadRequest,
			IsBase64Encoded: false,
			Body:            "Content-Type header not found",
		}, nil
	}
	if contentType != "application/octet-stream" {
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusBadRequest,
			IsBase64Encoded: false,
			Body:            "Content-Type header is not application/octet-stream",
		}, nil
	}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(request.Body))
	content, err := io.ReadAll(reader)
	if err != nil {
		return errorResponse(err)
	}
	err = s3Client.PutObject(ctx, bucket, key, content)
	if err != nil {
		return errorResponse(err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusAccepted,
	}, nil
}

func downloadFile(ctx context.Context, bucket, key string) (events.APIGatewayProxyResponse, error) {
	slog.Info("Downloading file from S3", "bucket", bucket, "key", key)
	reader, err := s3Client.GetObject(ctx, bucket, key)
	if err != nil {
		if errors.Is(err, api.ObjectNotFoundError) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, nil
		}
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)
	payload, err := io.ReadAll(reader)
	if err != nil {
		return errorResponse(err)
	}
	headers := map[string]string{
		"Content-Type": "application/octet-stream",
	}
	return events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         headers,
		IsBase64Encoded: true,
		Body:            base64.StdEncoding.EncodeToString(payload),
	}, nil
}

type ListFilesResponse struct {
	Files     []string `json:"files"`
	NextToken *string  `json:"nextToken,omitempty"`
}

func listFiles(ctx context.Context, bucket string, queryParams map[string]string) (events.APIGatewayProxyResponse, error) {
	slog.Info("Listing files from S3", "bucket", bucket)
	var nextToken *string
	if value, ok := queryParams["nextToken"]; ok {
		nextToken = &value
	}
	result, next, err := s3Client.ListObjects(ctx, bucket, nextToken)
	if err != nil {
		return errorResponse(err)
	}
	slog.Info(fmt.Sprintf("Fetched %d keys", len(result)))
	response := ListFilesResponse{
		Files:     result,
		NextToken: next,
	}
	body, err := json.Marshal(response)
	if err != nil {
		return errorResponse(err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:            string(body),
		IsBase64Encoded: false,
	}, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filePath := request.PathParameters["proxy"]
	bucket := os.Getenv("BUCKET_NAME")
	switch request.HTTPMethod {
	case "GET":
		var response events.APIGatewayProxyResponse
		var err error
		if request.Path == "/" {
			response, err = listFiles(ctx, bucket, request.QueryStringParameters)
		} else {
			response, err = downloadFile(ctx, bucket, filePath)
		}
		if err != nil {
			slog.Error("Request failed", "path", request.Path, "method", request.HTTPMethod, "error", err.Error())
			return errorResponse(err)
		}
		return response, nil
	case "POST":
		response, err := uploadFile(ctx, bucket, filePath, request)
		if err != nil {
			slog.Error("Request failed", "path", request.Path, "method", request.HTTPMethod, "error", err.Error())
			return errorResponse(err)
		}
		return response, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
	}, nil
}

func init() {
	pluginPath := os.Getenv("PLUGIN_PATH")
	p, err := plugin.Open(pluginPath)
	if err != nil {
		panic(err)
	}
	symbol, err := p.Lookup(S3ClientFactoryName)
	if err != nil {
		panic(err)
	}
	client, err := symbol.(func() (api.S3Client, error))()
	s3Client = client
	if err != nil {
		panic(err)
	}
}

func main() {
	lambda.Start(handler)
}
