package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"producer/internal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"google.golang.org/protobuf/proto"
)

var kinesisClient *kinesis.Client

const (
	UserEventCreated = 0
	UserEventUpdated = 1
	UserEventDeleted = 2
)

type UserEvent struct {
	EventType    string `json:"eventType"`
	EventId      string `json:"eventId"`
	UserId       string `json:"userId"`
	UserName     string `json:"userName"`
	UserEmail    string `json:"userEmail"`
	UserTimezone string `json:"userTimezone"`
}

func main() {
	var streamArn string
	var jsonData string
	flag.StringVar(&streamArn, "stream-arn", "", "Kinesis Stream ARN")
	flag.StringVar(&jsonData, "json", "", "JSON encoded input")
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	kinesisClient = kinesis.NewFromConfig(cfg)
	flag.Parse()
	var event UserEvent
	err = json.Unmarshal([]byte(jsonData), &event)
	if err != nil {
		panic(err)
	}
	var eventTypeCode int32
	switch event.EventType {
	case "CREATED":
		eventTypeCode = UserEventCreated
	case "UPDATED":
		eventTypeCode = UserEventUpdated
	case "DELETED":
		eventTypeCode = UserEventDeleted
	default:
		panic("Wrong event type")
	}
	eventType := internal.UserEventType(eventTypeCode)
	userEvent := internal.UserEvent_builder{
		EventType:    &eventType,
		EventId:      proto.String(event.EventId),
		UserId:       proto.String(event.UserId),
		UserName:     proto.String(event.UserName),
		UserEmail:    proto.String(event.UserEmail),
		UserTimezone: proto.String(event.UserTimezone),
	}.Build()
	msg, err := proto.Marshal(userEvent)
	if err != nil {
		panic(err)
	}
	out, err := kinesisClient.PutRecord(ctx, &kinesis.PutRecordInput{
		PartitionKey: aws.String(event.EventId),
		StreamARN:    aws.String(streamArn),
		Data:         msg,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Record with SequenceNumber: %s, ShardId: %s created\n", *out.SequenceNumber, *out.ShardId)
}
