//go:build integration

package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestAdapter_UploadDownload(t *testing.T) {
	ctx := context.Background()
	sdkCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(os.Getenv("S3_REGION")),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), "")))
	if err != nil {
		t.Fatalf("load s3 config: %v", err)
	}

	s3Client := awss3.NewFromConfig(sdkCfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT"))
		o.UsePathStyle = true
	})

	s3Adapter := NewAdapter(s3Client, os.Getenv("S3_BUCKET"))

	key := fmt.Sprintf("tests/upload-download-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() {
		_, err := s3Client.DeleteObject(ctx, &awss3.DeleteObjectInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(key),
		})
		if err != nil {
			t.Errorf("delete test object from s3: %v", err)
		}
	})

	expected := "Hello MinIO!"
	body := strings.NewReader(expected)
	contentType := "text/plain"

	err = s3Adapter.Upload(ctx, key, body, contentType)
	if err != nil {
		t.Fatalf("upload to s3: %v", err)
	}

	reader, downloadedContentType, err := s3Adapter.Download(ctx, key)
	if err != nil {
		t.Fatalf("download from s3: %v", err)
	}
	defer reader.Close()

	downloadbody, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read downloaded object: %v", err)
	}
	if string(downloadbody) != expected {
		t.Fatalf("content mismatch: expected %q, got %q", expected, string(downloadbody))
	}
	if downloadedContentType != contentType {
		t.Fatalf("content type mismatch: expected %q, got %q", contentType, downloadedContentType)
	}

	missingReader, _, err := s3Adapter.Download(ctx, key+"-missing")
	if err == nil {
		if missingReader != nil {
			missingReader.Close()
		}
		t.Fatal("expected error for missing object, got nil")
	}
}
