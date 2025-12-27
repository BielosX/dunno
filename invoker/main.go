package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

var initDurationRegexp *regexp.Regexp
var restoreDurationRegexp *regexp.Regexp
var lambdaClient *lambda.Client

func main() {
	initDurationRegexp = regexp.MustCompile("Init Duration:\\s*([\\d.]+)\\s*ms")
	restoreDurationRegexp = regexp.MustCompile("Restore Duration:\\s*([\\d.]+)\\s*ms")
	var functionName string
	var payload string
	flag.StringVar(&functionName, "name", "", "AWS Lambda function name")
	flag.StringVar(&payload, "payload", "", "AWS Lambda function input")
	flag.Parse()
	if functionName == "" {
		panic("name parameter required")
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err.Error())
	}
	lambdaClient = lambda.NewFromConfig(cfg)
	out, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		LogType:      types.LogTypeTail,
		Payload:      []byte(payload),
	})
	if err != nil {
		panic(err.Error())
	}
	resultBody := string(out.Payload)
	startupType := "Warm Start"
	var logResult string
	if out.LogResult != nil {
		r, err := base64.StdEncoding.DecodeString(*out.LogResult)
		if err != nil {
			panic(err.Error())
		}
		logResult = string(r)
		match := initDurationRegexp.FindStringSubmatch(logResult)
		if match != nil {
			startupType = fmt.Sprintf("Cold Start Duration: %s ms", match[1])
		}
		match = restoreDurationRegexp.FindStringSubmatch(logResult)
		if match != nil {
			startupType = fmt.Sprintf("Restore Duration: %s ms", match[1])
		}
	}
	fmt.Printf("%d | %s\n", out.StatusCode, startupType)
	if logResult != "" {
		fmt.Println(logResult)
	}
	fmt.Printf(resultBody)
}
