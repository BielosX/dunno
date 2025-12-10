package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"golang.org/x/sync/errgroup"
)

var lambdaClient *lambda.Client
var logsClient *cloudwatchlogs.Client
var initDurationRegexp *regexp.Regexp
var restoreDurationRegexp *regexp.Regexp
var logger *slog.Logger

func getStartupDuration(ctx context.Context, logGroup, logStreamName *string) (*float64, error) {
	logger.Info("Fetching log events", "logGroup", *logGroup, "logStream", *logStreamName)
	var nextToken *string
	for {
		getEventsOut, err := logsClient.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
			LogStreamName: logStreamName,
			LogGroupName:  logGroup,
			StartFromHead: aws.Bool(false),
		})
		if err != nil {
			var notFoundErr *cwTypes.ResourceNotFoundException
			if ok := errors.As(err, &notFoundErr); ok {
				logger.Info("LogStream or LogGroup not found, retrying")
				time.Sleep(1 * time.Second)
				continue
			}
			return nil, err
		}
		for _, event := range getEventsOut.Events {
			match := initDurationRegexp.FindStringSubmatch(*event.Message)
			if match != nil {
				duration, err := strconv.ParseFloat(match[1], 64)
				if err != nil {
					return nil, err
				}
				logger.Info("Initial Duration found", "value", duration)
				return &duration, nil
			}
			match = restoreDurationRegexp.FindStringSubmatch(*event.Message)
			if match != nil {
				duration, err := strconv.ParseFloat(match[1], 64)
				if err != nil {
					return nil, err
				}
				logger.Info("Restore Duration found", "value", duration)
				return &duration, nil
			}
		}
		nextToken = getEventsOut.NextBackwardToken
		if nextToken == nil {
			break
		}
	}
	return nil, nil
}

func invokeFunction(ctx context.Context, language, functionName, logGroup, arn *string) (*float64, error) {
	tagsOut, err := lambdaClient.ListTags(ctx, &lambda.ListTagsInput{
		Resource: arn,
	})
	if err != nil {
		logger.Error("unable to list tags", "error", err)
		os.Exit(1)
	}
	functionLanguage, ok := tagsOut.Tags["language"]
	if !ok || functionLanguage != *language {
		return nil, nil
	}
	logger.Info("Invoking function", "function", *functionName)
	invokeOut, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: functionName,
	})
	if err != nil {
		return nil, err
	}
	var logStreamName string
	err = json.Unmarshal(invokeOut.Payload, &logStreamName)
	if err != nil {
		return nil, err
	}
	var startupDuration *float64
	for startupDuration == nil {
		startupDuration, err = getStartupDuration(ctx, logGroup, &logStreamName)
		if err != nil {
			return nil, err
		}
		if startupDuration == nil {
			logger.Info("No Startup Duration found, retrying")
			time.Sleep(1 * time.Second)
		}
	}
	return startupDuration, nil
}

func main() {
	var functionVersion string
	var language string
	var logLevel string
	flag.StringVar(&functionVersion, "version", "", "Function version")
	flag.StringVar(&language, "language", "", "Function language")
	flag.StringVar(&logLevel, "loglevel", "error", "Function log level")
	flag.Parse()
	level := slog.Level(0)
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		panic(err)
	}
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.NewStandard(func(options *retry.StandardOptions) {
			options.MaxAttempts = 6
			options.MaxBackoff = time.Second * 10
		})
	}))
	if err != nil {
		logger.Error("unable to load SDK config", "error", err)
		os.Exit(1)
	}
	initDurationRegexp = regexp.MustCompile("Init Duration:\\s*([\\d.]+)\\s*ms")
	restoreDurationRegexp = regexp.MustCompile("Restore Duration:\\s*([\\d.]+)\\s*ms")
	lambdaClient = lambda.NewFromConfig(cfg)
	logsClient = cloudwatchlogs.NewFromConfig(cfg)

	count := 0
	maxValue := 0.0
	minValue := math.MaxFloat64
	averageValue := 0.0
	resultLock := sync.Mutex{}
	var marker *string
	for {
		result, err := lambdaClient.ListFunctions(ctx, &lambda.ListFunctionsInput{
			FunctionVersion: lambdaTypes.FunctionVersion(functionVersion),
			MaxItems:        aws.Int32(50),
			Marker:          marker,
		})
		if err != nil {
			logger.Error("unable to list functions", "error", err)
			os.Exit(1)
		}
		group, newCtx := errgroup.WithContext(ctx)
		for _, function := range result.Functions {
			group.Go(func() error {
				select {
				default:
					startupDuration, err := invokeFunction(newCtx,
						&language,
						function.FunctionName,
						function.LoggingConfig.LogGroup,
						function.FunctionArn)
					if err != nil {
						return err
					}
					if startupDuration != nil {
						resultLock.Lock()
						averageValue += *startupDuration
						if maxValue < *startupDuration {
							maxValue = *startupDuration
						}
						if minValue > *startupDuration {
							minValue = *startupDuration
						}
						count++
						resultLock.Unlock()
					}
					return nil
				case <-newCtx.Done():
					return nil
				}
			})
		}
		err = group.Wait()
		if err != nil {
			logger.Error("invocation failed", "error", err)
			os.Exit(1)
		}
		marker = result.NextMarker
		if marker == nil {
			break
		}
	}
	averageValue = averageValue / float64(count)
	fmt.Printf("Min: %f, Max: %f, Average: %f\n", minValue, maxValue, averageValue)
}
