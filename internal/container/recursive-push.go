package container

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/api/types/image"
)

type RecursivePush struct{}

func (rp *RecursivePush) PullImage(cf *DockerFactory, img string) (string, error) {
	// Get registry info from etcd
	localRegistryAddress, err := getLocalRegistryAddress()
	if err != nil {
		return img, fmt.Errorf("could not get local registry address from etcd: %v", err)
	}
	// Format the name for local image using local registry address
	img = strings.Join([]string{localRegistryAddress, img}, "/")

	// Try to pull first from local registry
	pullResp, err := cf.cli.ImagePull(cf.ctx, img, image.PullOptions{})
	if err != nil {
		return img, fmt.Errorf("could not pull image '%s': %v", img, err)
	}

	defer func(pullResp io.ReadCloser) {
		err := pullResp.Close()
		if err != nil {
			log.Printf("Could not close the docker image pull response\n")
		}
	}(pullResp)

	// This seems to be necessary to wait for the img to be pulled:
	_, _ = io.Copy(io.Discard, pullResp)
	log.Printf("Pulled image: %s\n", img)
	refreshedImages[img] = true

	return img, nil
}
