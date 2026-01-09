package main

import (
	"context"
	"flag"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"golang.org/x/sync/errgroup"
)

var sqsClient *sqs.Client

func main() {
	var queueUrl string
	var count int64
	flag.StringVar(&queueUrl, "queue-url", "", "SQS Url")
	flag.Int64Var(&count, "count", 1000, "Number of messages to send")
	flag.Parse()
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err.Error())
	}
	sqsClient = sqs.NewFromConfig(cfg)
	group, newCtx := errgroup.WithContext(ctx)
	counter := 0
	lock := sync.Mutex{}
	for i := int64(0); i < count; i++ {
		group.Go(func() error {
			_, err := sqsClient.SendMessage(newCtx, &sqs.SendMessageInput{
				QueueUrl:    aws.String(queueUrl),
				MessageBody: aws.String("Hello"),
			})
			lock.Lock()
			counter++
			fmt.Printf("\rMessages sent: %d", counter)
			lock.Unlock()
			return err
		})
	}
	err = group.Wait()
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("\rMessages sent: %d\n", counter)
}
