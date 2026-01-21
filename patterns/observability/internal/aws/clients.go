package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/v2/instrumentation/awsv2"
)

var Config aws.Config
var DynamoDbClient *dynamodb.Client
var CloudWatchClient *cloudwatch.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err.Error())
	}
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	Config = cfg
	DynamoDbClient = dynamodb.NewFromConfig(cfg)
	CloudWatchClient = cloudwatch.NewFromConfig(cfg)
}
