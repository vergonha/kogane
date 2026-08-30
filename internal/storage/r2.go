package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	bucket  string
	presign *s3.PresignClient
}

func New(
	bucket,
	accountID,
	accessKey,
	secretKey string,
) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	r2URL := fmt.Sprintf(
		"https://%s.r2.cloudflarestorage.com",
		accountID,
	)

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2URL)
	})

	return &Client{
		bucket:  bucket,
		presign: s3.NewPresignClient(s3Client),
	}, nil
}

func (c *Client) PresignGetObject(
	ctx context.Context,
	key string,
	expires time.Duration,
) (string, error) {
	req, err := c.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(expires),
	)
	if err != nil {
		return "", err
	}

	return req.URL, nil
}
