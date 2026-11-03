package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Adapter struct {
	client *awss3.Client
	bucket string
}

func NewAdapter(client *awss3.Client, bucket string) *Adapter {
	return &Adapter{
		client: client,
		bucket: bucket,
	}
}

func (a *Adapter) Upload(ctx context.Context, key string, body io.ReadSeeker, contentType string) error {
	input := &awss3.PutObjectInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}
	_, err := a.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("upload object to s3: %w", err)
	}
	return nil
}

func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	input := &awss3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	}
	output, err := a.client.GetObject(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("download object from s3: %w", err)
	}
	return output.Body, aws.ToString(output.ContentType), nil
}
