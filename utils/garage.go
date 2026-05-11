package utils

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client = nil
var s3Mutex sync.Mutex

// GetGarageClient return Garage client instance if exists. Otherwise, create and return it.
func GetGarageClient() (*s3.Client, error) {
	s3Mutex.Lock()
	defer s3Mutex.Unlock()

	// If client wasn't already created return it
	if s3Client != nil {
		return s3Client, nil
	}

	// TODO manage keys in better way (by means of config file)
	endpoint := "http://127.0.0.1:3900"
	accessKey := "GK845615fa124839a5db2d68043d8ab2ec"
	secretKey := "6142e5e266de127ce98333c124b6b708751804b54c1e2f697543a3475ba21902"
	region := "garage"

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
		o.UsePathStyle = true
	})

	return s3Client, nil
}
