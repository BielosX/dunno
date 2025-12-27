package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"golang.org/x/sync/errgroup"
)

var lambdaClient *lambda.Client
var initDurationRegexp *regexp.Regexp

type LambdaPayload struct {
	SleepSeconds int64 `json:"sleepSeconds"`
}

type LambdaResponse struct {
	LogStream string `json:"logStream"`
	EnvId     string `json:"envId"`
	RequestId string `json:"requestId"`
}

func invoke(ctx context.Context, name string, lambdaSleep int64) (*lambda.InvokeOutput, error) {
	payload := LambdaPayload{
		SleepSeconds: lambdaSleep,
	}
	encoded, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	out, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(name),
		LogType:      types.LogTypeTail,
		Payload:      encoded,
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	var functionName string
	var parallel int64
	var lambdaSleep int64
	flag.StringVar(&functionName, "function-name", "", "AWS Lambda function name")
	flag.Int64Var(&parallel, "parallel", 5, "Number of parallel invocations")
	flag.Int64Var(&lambdaSleep, "lambda-sleep", 5, "Value provided to lambda as sleepSeconds param")
	initDurationRegexp = regexp.MustCompile("Init Duration:\\s*([\\d.]+)\\s*ms")
	flag.Parse()
	if functionName == "" {
		panic("Function name must be provided")
	}
	if parallel <= 0 {
		panic("Parallel must be greater than zero")
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	lambdaClient = lambda.NewFromConfig(cfg)
	group, ctx := errgroup.WithContext(ctx)
	results := make([]string, parallel)
	for i := int64(0); i < parallel; i++ {
		group.Go(func() error {
			out, err := invoke(ctx, functionName, lambdaSleep)
			if err != nil {
				var apiErr smithy.APIError
				if errors.As(err, &apiErr) && apiErr.ErrorCode() == "TooManyRequestsException" {
					results[i] = "TooManyRequestsException"
					return nil
				}
				return err
			}
			if out.LogResult == nil {
				return errors.New("log result not available")
			}
			logResult, err := base64.StdEncoding.DecodeString(*out.LogResult)
			if err != nil {
				return err
			}
			var duration string
			match := initDurationRegexp.FindStringSubmatch(string(logResult))
			if match != nil {
				duration = fmt.Sprintf("Cold Start Duration: %s", match[1])
			} else {
				duration = "Warm Start"
			}
			var lambdaResponse LambdaResponse
			err = json.Unmarshal(out.Payload, &lambdaResponse)
			if err != nil {
				return err
			}
			results[i] = fmt.Sprintf("RequestId: %s EnvId: %s LogStream %s | %s",
				lambdaResponse.RequestId,
				lambdaResponse.EnvId,
				lambdaResponse.LogStream,
				duration)
			return nil
		})
	}
	err = group.Wait()
	if err != nil {
		panic(err.Error())
	}
	for i, result := range results {
		fmt.Printf("#%d %s\n", i, result)
	}
}
