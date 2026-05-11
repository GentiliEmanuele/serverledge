package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"io"

	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/serverledge-faas/serverledge/internal/cache"
	"github.com/serverledge-faas/serverledge/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
)

// Function describes a serverless function.
type Function struct {
	Name            string
	Runtime         string   // example: python314
	MemoryMB        int64    // MB
	CPUDemand       float64  // 1.0 -> 1 core
	MaxConcurrency  int16    // intra-container maximum concurrency
	Handler         string   // example: "module.function_name"
	TarFunctionCode string   // input is .tar
	CustomImage     string   // used if custom runtime is chosen
	SupportedArchs  []string // list of supported architectures by the runtime
	Signature       *Signature
}

const BucketName = "default-bucket"

func (f *Function) getEtcdKey() string {
	return getEtcdKey(f.Name)
}

func getEtcdKey(funcName string) string {
	return fmt.Sprintf("/function/%s", funcName)
}

func (f *Function) SupportsArch(arch string) bool {
	return slices.Contains(f.SupportedArchs, arch)
}

// GetFunction retrieves a Function given its name. If it doesn't exist, returns false
func GetFunction(name string) (*Function, bool) {

	val, found := getFromCache(name)
	if !found {
		// cache miss
		f, response := getFromGarage(name)
		if !response {
			return nil, false
		}
		//insert a new element to the cache
		cache.GetCacheInstance().Set(name, f, cache.DefaultExp)
		return f, true
	}

	return val, true

}

func (f *Function) String() string {
	return f.Name
}

func getFromCache(name string) (*Function, bool) {
	localCache := cache.GetCacheInstance()
	f, found := localCache.Get(name)
	if !found {
		return nil, false
	}
	//cache hit
	//return a safe copy of the function previously obtained
	function := *f.(*Function)
	return &function, true

}

// getFromGarage retrieve function infos from Garage
func getFromGarage(name string) (*Function, bool) {
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

// SaveToGarage registers the function to Garage
func (f *Function) SaveToGarage() error {
	// Get Garage client
	cli, err := utils.GetGarageClient()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(*f)
	if err != nil {
		return fmt.Errorf("could not marshal function: %v", err)
	}

	// In garage use /function/name as key
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

	// Add the function to the local cache
	cache.GetCacheInstance().Set(f.Name, f, cache.DefaultExp)

	return nil
}

// Delete removes a function from Garage and the local cache.
func (f *Function) Delete() error {
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

	// Remove the function from the local cache
	cache.GetCacheInstance().Delete(f.Name)

	return nil
}

func (f *Function) Equals(f2 *Function) bool {
	return (f == nil && f2 == nil) || (f.Name == f2.Name &&
		f.CustomImage == f2.CustomImage &&
		f.CPUDemand == f2.CPUDemand &&
		f.Runtime == f2.Runtime &&
		f.Handler == f2.Handler &&
		f.MemoryMB == f2.MemoryMB &&
		f.TarFunctionCode == f2.TarFunctionCode)
}

// Exists checks if the function is already saved to Etcd
func (f *Function) Exists() bool {
	savedFunction, ok := GetFunction(f.Name)
	return ok && f.Equals(savedFunction)
}

// GetAll returns all function names
func GetAll() ([]string, error) {
	return GetAllWithPrefix("/function")
}

// GetAllWithPrefix is used to get all /function or /workflow currently registered in etcd
func GetAllWithPrefix(prefix string) ([]string, error) {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	functions := make([]string, len(resp.Kvs))
	for i, s := range resp.Kvs {
		functions[i] = string(s.Key)[len(prefix+"/"):]
	}

	return functions, ctx.Err()
}
