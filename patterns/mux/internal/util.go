package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

type Response[T any] struct {
	Body         T
	StatusCode   int
	ErrorMessage *string
}

type Request[T any] struct {
	Body        T
	PathParams  map[string]string
	QueryParams map[string][]string
}

type Unit struct{}

type PathParams struct{}
type QueryParams struct{}

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
	handler func(request *Request[I]) Response[O]) *mux.Route {
	return router.HandleFunc(path, toHandleFunc(handler)).Methods(method)
}

func IsUnit(value any) bool {
	return reflect.TypeOf(value) == reflect.TypeOf(Unit{})
}

func toHandleFunc[I any, O any](handler func(request *Request[I]) Response[O]) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body I
		if !IsUnit(body) {
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		}
		response := handler(&Request[I]{
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
