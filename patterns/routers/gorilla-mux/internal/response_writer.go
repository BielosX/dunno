package internal

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type ApiGwResponseWriter struct {
	header     http.Header
	statusCode int
	body       []byte
}

func NewApiGwResponseWriter() *ApiGwResponseWriter {
	return &ApiGwResponseWriter{
		header:     make(http.Header),
		statusCode: 200,
		body:       nil,
	}
}

func (w *ApiGwResponseWriter) Header() http.Header {
	return w.header
}

func (w *ApiGwResponseWriter) Write(body []byte) (int, error) {
	w.body = body
	return len(body), nil
}

func (w *ApiGwResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *ApiGwResponseWriter) GetResponse() events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode:        w.statusCode,
		MultiValueHeaders: w.header,
		Body:              string(w.body),
	}
}
