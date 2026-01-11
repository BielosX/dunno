package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	dynamoTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
)

type Response[T any] struct {
	Body         *T
	StatusCode   int
	ErrorMessage *string
}

type Request[T any] struct {
	Context     context.Context
	Body        T
	PathParams  map[string]string
	QueryParams map[string][]string
}

type Unit struct{}

type PathParams struct{}
type QueryParams struct{}

func DecodeLastEvaluatedKey(key string) (map[string]dynamoTypes.AttributeValue, error) {
	decoded, err := io.ReadAll(base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(key)))
	if err != nil {
		return nil, err
	}
	var lastEvaluatedKey map[string]dynamoTypes.AttributeValue
	err = json.Unmarshal(decoded, &lastEvaluatedKey)
	if err != nil {
		return nil, err
	}
	return lastEvaluatedKey, nil
}

func ParseInt32(s string) (*int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil, err
	}
	value := int32(i)
	return &value, nil
}

func EncodeLastEvaluatedKey(key map[string]dynamoTypes.AttributeValue) (string, error) {
	result, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(result)
	return encoded, nil
}

func GetQueryParam[T any](r *Request[T], key string) string {
	value, ok := r.QueryParams[key]
	if !ok {
		return ""
	}
	return value[0]
}

func ErrorResponse[T any](err error, statusCode int) *Response[T] {
	message := err.Error()
	return &Response[T]{
		StatusCode:   statusCode,
		ErrorMessage: &message,
	}
}

func ServerError[T any](err error) *Response[T] {
	return ErrorResponse[T](err, http.StatusInternalServerError)
}

func SuccessResponse[T any](body *T) *Response[T] {
	return &Response[T]{
		Body:       body,
		StatusCode: http.StatusOK,
	}
}

func NotFoundResponse[T any]() *Response[T] {
	return &Response[T]{
		StatusCode: http.StatusNotFound,
	}
}

func GetFieldsByTagValue(s any, tag, value string) []reflect.StructField {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	var fields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tagValue, ok := f.Tag.Lookup(tag)
		if ok && tagValue == value {
			fields = append(fields, f)
		}
	}
	return fields
}

func WithPathParams(r *http.Request, params map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), PathParams{}, params)
	return r.WithContext(ctx)
}

func WithQueryParams(r *http.Request, params map[string][]string) *http.Request {
	ctx := context.WithValue(r.Context(), QueryParams{}, params)
	return r.WithContext(ctx)
}

func GetPathParams(r *http.Request) map[string]string {
	return r.Context().Value(PathParams{}).(map[string]string)
}

func GetQueryParams(r *http.Request) map[string][]string {
	return r.Context().Value(QueryParams{}).(map[string][]string)
}

func RegisterFunc[I any, O any](router *mux.Router,
	path, method string,
	handler func(request *Request[I]) *Response[O]) *mux.Route {
	return router.HandleFunc(path, toHandleFunc(handler)).Methods(method)
}

func RegisterFuncMatchContentType[I any, O any](router *mux.Router,
	path, method string,
	handler func(request *Request[I]) *Response[O],
	contentType string) *mux.Route {
	return RegisterFunc[I, O](router, path, method, handler).Headers("Content-Type", contentType)
}

func IsUnit(value any) bool {
	return reflect.TypeOf(value) == reflect.TypeOf(Unit{})
}

func toHandleFunc[I any, O any](handler func(request *Request[I]) *Response[O]) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body I
		if !IsUnit(body) {
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				err = fmt.Errorf("failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		}
		response := handler(&Request[I]{
			Context:     r.Context(),
			Body:        body,
			PathParams:  GetPathParams(r),
			QueryParams: GetQueryParams(r),
		})
		if response.StatusCode < 100 || response.StatusCode >= 600 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if response.ErrorMessage != nil {
			w.WriteHeader(response.StatusCode)
			_, _ = w.Write([]byte(*response.ErrorMessage))
			return
		}
		if !IsUnit(response.Body) {
			responseBody, err := json.Marshal(response.Body)
			if err != nil {
				err = fmt.Errorf("failed to encode response body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(response.StatusCode)
			_, _ = w.Write(responseBody)
			return
		} else {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
}
