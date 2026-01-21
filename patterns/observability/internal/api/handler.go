package api

import (
	"dunno/internal/api/books"
	"dunno/internal/api/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var LambdaHandler Handler

func Init() {
	config.Load()
	rootRouter := chi.NewRouter()
	rootRouter.Use(middleware.Logger)
	rootRouter.Use(endpointDurationMiddleware(config.ApiConfig.MetricsNamespace, "requestDuration"))
	rootRouter.Mount("/books", books.NewRouter())
	LambdaHandler = NewHandler(rootRouter)
}
