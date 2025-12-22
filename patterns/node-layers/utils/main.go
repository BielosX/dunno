package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

var lambdaClient *lambda.Client

const (
	getLatestLayerVersionCommand      = "get-latest-layer-version"
	deleteLayerCommand                = "delete-layer"
	getLatestLayerVersionArnOutput    = "arn"
	getLatestLayerVersionNumberOutput = "number"
)

func deleteLayer(ctx context.Context, layerName string) {
	if layerName == "" {
		println("No layer name provided")
		os.Exit(1)
	}
	paginator := lambda.NewListLayerVersionsPaginator(lambdaClient, &lambda.ListLayerVersionsInput{
		LayerName: aws.String(layerName),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		for _, layerVersion := range page.LayerVersions {
			_, err = lambdaClient.DeleteLayerVersion(ctx, &lambda.DeleteLayerVersionInput{
				LayerName:     aws.String(layerName),
				VersionNumber: aws.Int64(layerVersion.Version),
			})
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
			println("Deleted Layer Version:", *layerVersion.LayerVersionArn)
		}
	}
}

func getLatestLayerVersion(ctx context.Context, layerName string, output string) {
	if layerName == "" {
		println("No layer name provided")
		os.Exit(1)
	}
	paginator := lambda.NewListLayerVersionsPaginator(lambdaClient, &lambda.ListLayerVersionsInput{
		LayerName: aws.String(layerName),
	})
	latestVersion := int64(0)
	var lastestVersionArn *string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		for _, version := range page.LayerVersions {
			if version.Version > latestVersion {
				latestVersion = version.Version
				lastestVersionArn = version.LayerVersionArn
			}
		}
	}
	if lastestVersionArn != nil {
		switch output {
		case getLatestLayerVersionArnOutput:
			fmt.Print(*lastestVersionArn)
			break
		case getLatestLayerVersionNumberOutput:
			fmt.Print(latestVersion)
			break
		default:
			fmt.Printf("%s or %s output expected", getLatestLayerVersionArnOutput, getLatestLayerVersionNumberOutput)
			os.Exit(1)
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flagSet := flag.NewFlagSet("flags", flag.PanicOnError)
	args := os.Args[1:]
	if len(args) == 0 {
		println("Command expected")
		os.Exit(1)
	}
	command := args[0]
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	lambdaClient = lambda.NewFromConfig(cfg)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	switch command {
	case getLatestLayerVersionCommand:
		name := flagSet.String("layer-name", "", "Layer Name")
		output := flagSet.String("out", getLatestLayerVersionArnOutput, "ARN or Number")
		must(flagSet.Parse(args[1:]))
		getLatestLayerVersion(ctx, *name, *output)
		break
	case deleteLayerCommand:
		name := flagSet.String("layer-name", "", "Layer Name")
		must(flagSet.Parse(args[1:]))
		deleteLayer(ctx, *name)
		break
	default:
		println("Unknown command %s", command)
		os.Exit(1)
	}
}
