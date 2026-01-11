package main

import (
	"dunno/internal"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(internal.HandleRequest)
}
