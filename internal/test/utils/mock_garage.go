package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type GarageFakeStorage struct {
	mu    sync.RWMutex
	store map[string][]byte
}

func NewGarageFakeStorage() *GarageFakeStorage {
	return &GarageFakeStorage{
		store: make(map[string][]byte),
	}
}

// PutObject simulate the saving in Garage
func (f *GarageFakeStorage) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if params.Key == nil || params.Body == nil {
		return nil, fmt.Errorf("invalid parameters")
	}

	bodyBytes, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}

	f.store[*params.Key] = bodyBytes
	return &s3.PutObjectOutput{}, nil
}

// GetObject simulate the reading from Garage
func (f *GarageFakeStorage) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if params.Key == nil {
		return nil, fmt.Errorf("invalid key")
	}

	data, exists := f.store[*params.Key]
	if !exists {
		return nil, fmt.Errorf("NoSuchKey: the specified key does not exist")
	}

	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(data)),
	}, nil
}

func (f *GarageFakeStorage) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if params.Key == nil {
		return nil, fmt.Errorf("invalid key")
	}

	delete(f.store, *params.Key)

	return &s3.DeleteObjectOutput{}, nil
}
