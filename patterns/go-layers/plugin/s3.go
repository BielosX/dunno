package main

import (
	"bytes"
	"context"
	"dunno/api"
	"errors"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type s3Client struct {
	client *s3.Client
}

//goland:noinspection GoUnusedExportedFunction
func NewS3Client() (api.S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &s3Client{
		client: client,
	}, err
}

func (c *s3Client) PutObject(ctx context.Context, bucket, key string, data []byte) error {
	slog.Info("PutObject called", "bucket", bucket, "key", key)
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *s3Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	slog.Info("GetObject called", "bucket", bucket, "key", key)
	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchKey" {
				return nil, api.ObjectNotFoundError
			}
		}
		return nil, err
	}
	return out.Body, nil
}

func (c *s3Client) ListObjects(ctx context.Context, bucket string, nextToken *string) ([]string, *string, error) {
	slog.Info("ListObjects called", "bucket", bucket, "nextToken", nextToken)
	var result = make([]string, 0)
	out, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucket),
		ContinuationToken: nextToken,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, obj := range out.Contents {
		result = append(result, *obj.Key)
	}
	return result, out.NextContinuationToken, nil
}
