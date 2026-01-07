package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"golang.org/x/sync/errgroup"
)

var lambdaClient *lambda.Client
var initDurationRegexp *regexp.Regexp
var restoreDurationRegexp *regexp.Regexp
var logger *slog.Logger

func calcMedian(values []float64) float64 {
	size := len(values)
	sort.Float64s(values)
	if size == 0 {
		return math.NaN()
	}
	if size%2 == 0 {
		index := size >> 1
		return (values[index] + values[index-1]) / 2.0
	}
	return values[size>>1]
}

func invokeFunction(ctx context.Context, language, functionName, alias, arn *string) (*float64, error) {
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
	var qualifier *string
	if *alias != "" {
		qualifier = alias
	}
	invokeOut, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: arn,
		LogType:      lambdaTypes.LogTypeTail,
		Qualifier:    qualifier,
	})
	if err != nil {
		return nil, err
	}
	if invokeOut.LogResult != nil {
		bytes, err := base64.StdEncoding.DecodeString(*invokeOut.LogResult)
		logResult := string(bytes)
		if err != nil {
			return nil, err
		}
		match := initDurationRegexp.FindStringSubmatch(logResult)
		if match != nil {
			duration, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return nil, err
			}
			logger.Info("Initial Duration found", "value", duration)
			return &duration, nil
		}
		match = restoreDurationRegexp.FindStringSubmatch(logResult)
		if match != nil {
			duration, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return nil, err
			}
			logger.Info("Restore Duration found", "value", duration)
			return &duration, nil
		}
		logger.Warn("Initial or Restore Duration not found", "name", *functionName)
	} else {
		logger.Warn("Log Result not available", "name", *functionName)
	}
	return nil, nil
}

func main() {
	var language string
	var logLevel string
	var alias string
	var parallel int64
	flag.StringVar(&language, "language", "", "Function language")
	flag.StringVar(&logLevel, "log-level", "info", "Function log level")
	flag.StringVar(&alias, "alias", "", "Version Alias")
	flag.Int64Var(&parallel, "parallel", -1, "Goroutine parallel invocations")
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

	count := 0
	maxValue := 0.0
	minValue := math.MaxFloat64
	averageValue := 0.0
	var results []float64
	resultLock := sync.Mutex{}
	paginator := lambda.NewListFunctionsPaginator(lambdaClient, &lambda.ListFunctionsInput{
		MaxItems: aws.Int32(50),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			logger.Error("unable to list functions", "error", err)
			os.Exit(1)
		}
		group, newCtx := errgroup.WithContext(ctx)
		if parallel > 0 {
			group.SetLimit(int(parallel))
		}
		for _, function := range page.Functions {
			logger.Info("Found function", "arn", *function.FunctionArn)
			group.Go(func() error {
				select {
				default:
					startupDuration, err := invokeFunction(newCtx,
						&language,
						function.FunctionName,
						&alias,
						function.FunctionArn)
					if err != nil {
						return err
					}
					if startupDuration != nil {
						resultLock.Lock()
						results = append(results, *startupDuration)
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
	}
	averageValue = averageValue / float64(count)
	median := calcMedian(results)
	fmt.Printf("Min: %f, Max: %f, Median: %f, Average: %f\n", minValue, maxValue, median, averageValue)
}
