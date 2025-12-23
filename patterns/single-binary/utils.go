package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Request events.APIGatewayV2HTTPRequest
type Response events.APIGatewayV2HTTPResponse

func InternalServerError(err error) Response {
	return Response{
		StatusCode: http.StatusInternalServerError,
		Body:       err.Error(),
	}
}

func getOptional[K comparable, V any](m map[K]V, key K) *V {
	value, ok := m[key]
	if !ok {
		return nil
	}
	return &value
}

func parseInt32(s string) (int32, error) {
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

func getHeader(headers map[string]string, headerKey string) *string {
	for key, value := range headers {
		if strings.ToLower(key) == strings.ToLower(headerKey) {
			return &value
		}
	}
	return nil
}

func convertLastEvaluatedKey(m map[string]types.AttributeValue) (*string, error) {
	if m != nil {
		out, err := attributevalue.MarshalMapJSON(m)
		if err != nil {
			return nil, nil
		}
		result := base64.RawURLEncoding.EncodeToString(out)
		return &result, nil
	}
	return nil, nil
}

func unmarshallBody[T any](request *Request) (*T, error) {
	var result T
	if request.IsBase64Encoded {
		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(request.Body))
		err := json.NewDecoder(reader).Decode(&result)
		if err != nil {
			return nil, err
		}
	} else {
		err := json.Unmarshal([]byte(request.Body), &result)
		if err != nil {
			return nil, err
		}
	}
	return &result, nil
}
