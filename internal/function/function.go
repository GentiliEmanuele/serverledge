package function

import (
	"fmt"
	"slices"

	"github.com/serverledge-faas/serverledge/internal/cache"
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
		f, response := GetStorage().Get(name)
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

// SaveFunction registers the function to the specified storage
func (f *Function) SaveFunction() error {
	err := GetStorage().Save(f)

	if err != nil {
		return fmt.Errorf("failed to save the function: %v", err)
	}

	// Add the function to the local cache
	cache.GetCacheInstance().Set(f.Name, f, cache.DefaultExp)

	return nil
}

// Delete removes a function from specified storage and from local cache.
func (f *Function) Delete() error {
	err := GetStorage().Delete(f)

	if err != nil {
		return fmt.Errorf("failed to delete the function: %v", err)
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
	return GetStorage().GetAll()
}
