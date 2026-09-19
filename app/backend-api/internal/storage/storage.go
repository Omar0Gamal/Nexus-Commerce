package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3        *s3.Client
	bucket    string
	publicURL string
}

func New(accountID, accessKey, secretKey, bucket, publicURL string) (*Client, error) {
	if accountID == "" {
		return nil, fmt.Errorf("storage: accountID is required")
	}
	if accessKey == "" {
		return nil, fmt.Errorf("storage: accessKey is required")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("storage: secretKey is required")
	}
	if bucket == "" {
		return nil, fmt.Errorf("storage: bucket is required")
	}
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	return NewWithEndpoint(endpoint, accessKey, secretKey, bucket, publicURL)
}

// Configures client for tests pointing at a local MinIO instance.
func NewWithEndpoint(endpoint, accessKey, secretKey, bucket, publicURL string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &Client{
		s3:        s3Client,
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

func (c *Client) Upload(ctx context.Context, key, contentType string, body io.Reader) (string, error) {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 put object: %w", err)
	}
	return c.PublicURL(key), nil
}

// Leading slashes in key are stripped to avoid double-slash URLs.
func (c *Client) PublicURL(key string) string {
	return c.publicURL + "/" + strings.TrimPrefix(key, "/")
}

func (c *Client) KeyFromURL(url string) string {
	prefix := c.publicURL + "/"
	if !strings.HasPrefix(url, prefix) {
		return ""
	}
	return strings.TrimPrefix(url, prefix)
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete object: %w", err)
	}
	return nil
}

func (c *Client) PresignUpload(ctx context.Context, key string, contentType string, expiresInSeconds int64) (string, string, error) {
	presignClient := s3.NewPresignClient(c.s3)
	
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expiresInSeconds) * time.Second
	})

	if err != nil {
		return "", "", fmt.Errorf("s3 presign upload: %w", err)
	}

	return req.URL, c.PublicURL(key), nil
}

func DetectContentType(fileHeader []byte) string {
	return http.DetectContentType(fileHeader)
}
