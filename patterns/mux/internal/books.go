package internal

import (
	"dunno/internal/global"
	"net/http"

	_ "dunno/internal/global" // Init Router first

	"go.uber.org/zap"
)

type BookResponse struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

func GetBook(r *Request[Unit]) Response[BookResponse] {
	bookId := r.PathParams["id"]
	global.Logger.Info("Fetching book with ID", zap.String("id", bookId))
	return Response[BookResponse]{
		Body: BookResponse{
			Id:    "1234",
			Title: "The Book",
		},
		StatusCode: http.StatusOK,
	}
}

func init() {
	RegisterFunc(global.Router, "/books/{id}", "GET", GetBook)
}
