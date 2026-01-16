package api

import (
	"context"
	"dunno/api/books"
	"dunno/api/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var rootRouter chi.Router

func init() {
	rootRouter = chi.NewRouter()
	rootRouter.Use(middleware.Logger)
	rootRouter.Mount("/books", books.Router)
}

func Handler(ctx context.Context, event Request) (Response, error) {
	log.Logger.Infof("Received ApiGateway event: %v", event)
	writer := NewResponseWriter()
	request, err := MapRequest(ctx, &event)
	if err != nil {
		return Response{}, err
	}
	rootRouter.ServeHTTP(writer, request)
	return writer.ToResponse(), nil
}
