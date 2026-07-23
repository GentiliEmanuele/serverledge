package utils

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/serverledge-faas/serverledge/internal/test/utils"

	sledgeConfig "github.com/serverledge-faas/serverledge/internal/config"
)

type GarageClientInterface interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

var s3Client GarageClientInterface = nil
var s3Mutex sync.Mutex

var UseFakeGarage = false // Set to true only for testing

// GetGarageClient return Garage client instance if exists. Otherwise, create and return it.
func GetGarageClient() (GarageClientInterface, error) {
	s3Mutex.Lock()
	defer s3Mutex.Unlock()

	// If client wasn't already created return it
	if s3Client != nil {
		return s3Client, nil
	}

	endpoint := sledgeConfig.GetString(sledgeConfig.GARAGE_ENDPOINT, "http://127.0.0.1:3900")
	accessKey := sledgeConfig.GetString(sledgeConfig.GARAGE_ACCESS_KEY, "")
	secretKey := sledgeConfig.GetString(sledgeConfig.GARAGE_SECRET_KEY, "")
	region := sledgeConfig.GetString(sledgeConfig.GARAGE_REGION, "garage")

	if UseFakeGarage {
		s3Client = utils.NewGarageFakeStorage()
		return s3Client, nil
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
		o.UsePathStyle = true
	})

	return s3Client, nil
}
