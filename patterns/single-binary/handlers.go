package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type Handlers struct {
	handlers map[string]interface{}
}

func newHandlers() *Handlers {
	return &Handlers{
		handlers: make(map[string]interface{}),
	}
}

func (h *Handlers) register(name string, handler interface{}) {
	h.handlers[name] = handler
}

func (h *Handlers) run() {
	name := os.Getenv("_HANDLER")
	handler, ok := h.handlers[name]
	if !ok {
		panic(fmt.Sprintf("no handler found with name %s", name))
	}
	lambda.Start(handler)
}

var handlers *Handlers
