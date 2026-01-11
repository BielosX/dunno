package internal

import (
	"context"
	"dunno/internal/global"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func HandleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := request.RequestContext.HTTP.Method
	path := request.RawPath
	global.Logger.Info(ctx, "Request received",
		zap.String("path", request.RawPath),
		zap.String("method", method),
		zap.String("body", request.Body))
	httpRequest, err := http.NewRequestWithContext(ctx, method, request.RawPath, strings.NewReader(request.Body))
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	headers := make(map[string][]string)
	for key, value := range request.Headers {
		parts := strings.Split(value, ",")
		headers[key] = parts
	}
	httpRequest.Header = headers

	var routeMatch mux.RouteMatch
	if global.Router.Match(httpRequest, &routeMatch) {
		global.Logger.Info(ctx, "Route matched", zap.String("path", path))
		responseWriter := NewApiGwResponseWriter()
		httpRequest = WithPathParams(httpRequest, routeMatch.Vars)
		httpRequest = WithQueryParams(httpRequest, request.QueryStringParameters)
		routeMatch.Handler.ServeHTTP(responseWriter, httpRequest)
		return responseWriter.GetResponse(), nil
	}
	global.Logger.Info(ctx, "Route not matched", zap.String("path", path))
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNotFound,
	}, nil
}
