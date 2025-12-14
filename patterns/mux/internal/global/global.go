package global

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger
var Router *mux.Router

func init() {
	prodLogger, _ := zap.NewProduction()
	Logger = prodLogger.Sugar()
	Router = mux.NewRouter()
}
