package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type Request events.APIGatewayV2HTTPRequest
type Response events.APIGatewayV2HTTPResponse

type ResponseWriter struct {
	headers       http.Header
	status        int
	body          []byte
	headerWritten bool
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		headers:       make(http.Header),
		body:          make([]byte, 0),
		headerWritten: false,
	}
}

func (w *ResponseWriter) Header() http.Header {
	return w.headers
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.headerWritten = true
}

func (w *ResponseWriter) Write(bytes []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	w.body = append(w.body, bytes...)
	return len(bytes), nil
}

func (w *ResponseWriter) ToResponse() Response {
	return Response{
		StatusCode:        w.status,
		Body:              string(w.body),
		MultiValueHeaders: w.headers,
	}
}

func MapRequest(ctx context.Context, r *Request) (*http.Request, error) {
	method := r.RequestContext.HTTP.Method
	url := r.RawPath
	if r.RawQueryString != "" {
		url = fmt.Sprintf("%s?%s", r.RawPath, r.RawQueryString)
	}
	request, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(r.Body))
	if err != nil {
		return nil, err
	}
	for key, values := range r.Headers {
		headerValues := strings.Split(values, ",")
		for i := range headerValues {
			request.Header.Set(key, strings.TrimSpace(headerValues[i]))
		}
	}
	return request, nil
}
