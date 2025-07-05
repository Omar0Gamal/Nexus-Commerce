package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client wraps an S3-compatible client pointed at a Cloudflare R2 bucket.
type Client struct {
	s3        *s3.Client
	bucket    string
	publicURL string // base URL for public object access, e.g. "https://pub-xxx.r2.dev"
}

// New creates a Client for the given Cloudflare R2 account.
// endpoint defaults to the standard R2 endpoint for the account.
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

// Useful for tests that point at a local MinIO instance.
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

// Upload stores body under key in the bucket with the given content-type and
// returns the public access URL.
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

// PublicURL returns the public access URL for the given storage key.
// Leading slashes in key are stripped to avoid double-slash URLs.
func (c *Client) PublicURL(key string) string {
	return c.publicURL + "/" + strings.TrimPrefix(key, "/")
}

// KeyFromURL extracts the storage key from a full public URL.
// Returns an empty string if the URL does not match this client's publicURL base.
func (c *Client) KeyFromURL(url string) string {
	prefix := c.publicURL + "/"
	if !strings.HasPrefix(url, prefix) {
		return ""
	}
	return strings.TrimPrefix(url, prefix)
}

// Delete removes the object at key from the bucket.
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

// DetectContentType sniffs the MIME type from the first 512 bytes.
func DetectContentType(data []byte) string {
	return http.DetectContentType(data)
}
