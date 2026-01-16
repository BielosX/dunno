package api

import (
	"dunno/api/books"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var rootRouter chi.Router
var LambdaHandler Handler

func init() {
	rootRouter = chi.NewRouter()
	rootRouter.Use(middleware.Logger)
	rootRouter.Mount("/books", books.Router)
	LambdaHandler = NewHandler(rootRouter)
}
