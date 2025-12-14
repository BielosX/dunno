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

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	global.Logger.Info(ctx, "Request received",
		zap.String("path", request.Path),
		zap.String("method", request.HTTPMethod))
	httpRequest, err := http.NewRequestWithContext(ctx, request.HTTPMethod, request.Path, strings.NewReader(request.Body))
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	httpRequest.Header = request.MultiValueHeaders

	var routeMatch mux.RouteMatch
	if global.Router.Match(httpRequest, &routeMatch) {
		responseWriter := NewApiGwResponseWriter()
		httpRequest = WithPathParams(httpRequest, routeMatch.Vars)
		httpRequest = WithQueryParams(httpRequest, request.MultiValueQueryStringParameters)
		routeMatch.Handler.ServeHTTP(responseWriter, httpRequest)
		return responseWriter.GetResponse(), nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
	}, nil
}
