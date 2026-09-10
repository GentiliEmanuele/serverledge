package function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/serverledge-faas/serverledge/utils"
)

type GarageStorage struct{}

const BucketName = "default-bucket"

func (gs *GarageStorage) Get(name string) (*Function, bool) {
	// Get Garage client
	cli, err := utils.GetGarageClient()
	if err != nil {
		return nil, false
	}

	// Create a context
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	// Format key
	key := fmt.Sprintf("function/%s", name)

	// Retrieve function object from Garage
	output, err := cli.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, false
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(output.Body)

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, false
	}

	var f Function
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, false
	}

	return &f, true
}

func (gs *GarageStorage) Save(f *Function) error {
	// Get Garage client
	cli, err := utils.GetGarageClient()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(*f)
	if err != nil {
		return fmt.Errorf("could not marshal function: %v", err)
	}

	// In garage use function/name as key
	key := fmt.Sprintf("function/%s", f.Name)

	// Write the function code in Garage
	_, err = cli.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(payload),
	})

	if err != nil {
		return fmt.Errorf("failed to save to Garage: %v", err)
	}

	return nil
}

func (gs *GarageStorage) Delete(f *Function) error {
	// Get Garage client
	cli, err := utils.GetGarageClient()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("function/%s", f.Name)
	_, err = cli.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from Garage: %v", err)
	}

	return nil
}

func (gs *GarageStorage) GetAll() ([]string, error) {
	return GetStorage().GetAllWithPrefix("function/")
}

func (gs *GarageStorage) GetAllWithPrefix(prefix string) ([]string, error) {
	// Get the garage client
	cli, err := utils.GetGarageClient()
	if err != nil {
		return nil, err
	}

	// Create a context
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	// Prepare the request
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(BucketName),
		Prefix: aws.String(prefix),
	}

	// Init the paginator
	paginator := s3.NewListObjectsV2Paginator(cli, params)

	keys := make([]string, 0)

	// Iterate on the pages
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			keys = append(keys, (*obj.Key)[len(prefix):])
		}
	}

	return keys, ctx.Err()
}
