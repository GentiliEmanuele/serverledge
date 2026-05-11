package utils

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client = nil
var s3Mutex sync.Mutex

func GetGarageClient() (*s3.Client, error) {
	s3Mutex.Lock()
	defer s3Mutex.Unlock()

	// If client was already created return it
	if s3Client != nil {
		return s3Client, nil
	}

	endpoint := "http://127.0.0.1:3900"
	accessKey := os.Getenv("GARAGE_DEFAULT_ACCESS_KEY")
	secretKey := os.Getenv("GARAGE_DEFAULT_SECRET_KEY")
	region := "garage"

	// Check if the key values correctly are retrieve from env variables
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("garage credentials not set in environment variables")
	}

	// Load base configuration, without resolver
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)

	if err != nil {
		return nil, err
	}

	// Configure endpoint
	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Return immutable host name
	})

	return s3Client, nil
}
