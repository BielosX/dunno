package api

import (
	"context"
	"errors"
	"io"
)

var ObjectNotFoundError = errors.New("object not found")

type S3Client interface {
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	PutObject(ctx context.Context, bucket, key string, data []byte) error
	ListObjects(ctx context.Context, bucket string, nextToken *string) ([]string, *string, error)
}
